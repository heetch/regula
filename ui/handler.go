package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/heetch/regula"
	reghttp "github.com/heetch/regula/http"
	regrule "github.com/heetch/regula/rule"
	"github.com/heetch/regula/rule/sexpr"
	"github.com/heetch/regula/store"
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// NewHandler creates a http handler serving the UI application and the UI backend.
func NewHandler(service store.RulesetService, fs http.FileSystem) http.Handler {
	var mux http.ServeMux

	// internal API
	mux.Handle("/i/", http.StripPrefix("/i", newInternalHandler(service)))

	// static files
	h := http.FileServer(fs)
	mux.Handle("/css/", h)
	mux.Handle("/js/", h)
	mux.Handle("/fonts/", h)

	// catch all url that deleguates the routing to the front app router
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := fs.Open("index.html")
		if err != nil {
			writeError(w, r, err, http.StatusInternalServerError)
			return
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		if err != nil {
			writeError(w, r, err, http.StatusInternalServerError)
			return
		}
	})

	return &mux
}

type uiError struct {
	Err string `json:"error"`
}

// writeError writes an error to the http response in JSON format.
func writeError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// Prepare log.
	logger := reghttp.LoggerFromRequest(r).With().
		Err(err).
		Int("status", code).
		Logger()

	// Hide error from client if it's internal.
	if code == http.StatusInternalServerError {
		logger.Error().Msg("unexpected http error")
		err = errInternal
	} else {
		logger.Debug().Msg("http error")
	}

	v, ok := err.(RuleError)
	if ok {
		reghttp.EncodeJSON(w, r, &v, code)
	} else {
		reghttp.EncodeJSON(w, r, &uiError{Err: err.Error()}, code)
	}
}

// handler serving the UI internal API.
type internalHandler struct {
	service store.RulesetService
}

func newInternalHandler(service store.RulesetService) http.Handler {
	h := internalHandler{
		service: service,
	}
	var mux http.ServeMux

	// router for the internal API
	mux.Handle("/rulesets/", h.rulesetsHandler())

	return &mux
}

// handleNewRulesetRequest consumes a POST to the ruleset endpoint and
// attempts to create a new Ruleset from that data.
func (h *internalHandler) handleNewRulesetRequest(w http.ResponseWriter, r *http.Request) {
	nrr := &newRulesetRequest{}
	err := json.NewDecoder(r.Body).Decode(nrr)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}
	rs, err := newRulesetFromRequest(nrr)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}
	_, err = h.service.Put(r.Context(), nrr.Path, rs)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}
	reghttp.EncodeJSON(w, r, nil, http.StatusCreated)

}

// handleListRequest attempts to return a list of Rulesets based on the data provided in a GET request to the ruleset endpoint.
func (h *internalHandler) handleListRequest(w http.ResponseWriter, r *http.Request) {
	type ruleset struct {
		Path string `json:"path"`
	}

	type response struct {
		Rulesets []ruleset `json:"rulesets"`
	}

	var resp response
	opt := store.ListOptions{
		Limit:     100,
		PathsOnly: true,
	}

	// run the loop at least once, no matter of the value of token
	for i := 0; i == 0 || opt.ContinueToken != ""; i++ {
		list, err := h.service.List(r.Context(), "", &opt)
		if err != nil {
			writeError(w, r, err, http.StatusInternalServerError)
			return
		}

		opt.ContinueToken = list.Continue
		for _, rs := range list.Entries {
			resp.Rulesets = append(resp.Rulesets, ruleset{Path: rs.Path})
		}
	}

	// set the slice to an empty slice to
	// prevent sending null if the list is empty.
	if resp.Rulesets == nil {
		resp.Rulesets = []ruleset{}
	}

	reghttp.EncodeJSON(w, r, &resp, http.StatusOK)
}

// Returns an http handler that lists all existing rulesets paths.
func (h *internalHandler) rulesetsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			h.handleNewRulesetRequest(w, r)
		case "GET":
			h.handleListRequest(w, r)
		}
	})
}

type param map[string]string

type signature struct {
	Params     []param `json:"params"`
	ReturnType string
}

type rule struct {
	SExpr       string `json:"sExpr"`
	ReturnValue string `json:"returnValue"`
}

// newRulesetRequest is the unmarshaled form a new ruleset request.
type newRulesetRequest struct {
	Path      string    `json:"path"`
	Signature signature `json:"signature"`
	Rules     []rule    `json:"rules"`
}

// convertParams takes a slice of param, unmarshalled from a
// newRulesetRequest, and returns an equivalent sexpr.Parameters map.
func convertParams(input []param) (sexpr.Parameters, error) {
	parm := make(sexpr.Parameters)
	for i, p := range input {
		name, found := p["name"]
		if !found {
			return nil, fmt.Errorf("parameter %d has no name", i)
		}
		v, found := p["type"]
		if !found {
			return nil, fmt.Errorf("parameter %d (%s) has no type", i, name)
		}
		t, err := regrule.TypeFromName(v)
		if err != nil {
			return nil, err
		}
		parm[name] = t
	}
	return parm, nil
}

// newRulesetFromRequest takes a newRulesetRequest and returns the
// equivalent Regula.Ruleset.
func newRulesetFromRequest(nrr *newRulesetRequest) (*regula.Ruleset, error) {
	var rules []*regrule.Rule
	parm, err := convertParams(nrr.Signature.Params)
	if err != nil {
		return nil, err
	}

	for n, rule := range nrr.Rules {
		p := sexpr.NewParser(bytes.NewBufferString(rule.SExpr))
		expr, err := p.Parse(parm)
		if err != nil {
			return nil, newRuleError(n+1, err)
		}
		val, err := makeValue(nrr.Signature.ReturnType, rule.ReturnValue)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &regrule.Rule{
			Expr:   expr,
			Result: val,
		})

	}

	return makeRuleset(nrr.Signature.ReturnType, rules...)
}

// makeValue takes a string representation of a value and its type, and returns the appropriate Value expression from the Regula rule library.
func makeValue(returnType, returnValue string) (*regrule.Value, error) {
	switch returnType {
	case "string":
		return regrule.StringValue(returnValue), nil
	case "bool":
		b, err := strconv.ParseBool(returnValue)
		if err != nil {
			return nil, err
		}
		return regrule.BoolValue(b), nil
	case "int64":
		i, err := strconv.ParseInt(returnValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return regrule.Int64Value(i), nil
	case "float64":
		f, err := strconv.ParseFloat(returnValue, 64)
		if err != nil {
			return nil, err
		}
		return regrule.Float64Value(f), nil
	}
	return nil, fmt.Errorf("Invalid return type %q", returnType)
}

func makeRuleset(returnType string, rules ...*regrule.Rule) (*regula.Ruleset, error) {
	switch returnType {
	case "string":
		return regula.NewStringRuleset(rules...)
	case "bool":
		return regula.NewBoolRuleset(rules...)
	case "int64":
		return regula.NewInt64Ruleset(rules...)
	case "float64":
		return regula.NewFloat64Ruleset(rules...)
	}
	return nil, fmt.Errorf("Invalid return type %q", returnType)
}

// A RuleError represents an error found within a single Rule within a
// Ruleset.  Currently these errors are all results of attempting to
// parse the rule using the symbolic expression parser.
type RuleError struct {
	ruleNum int
	err     error
}

// newRuleError creates a new RuleError for the provided ruleNum.  This RuleError will wrap the provided qerr.
func newRuleError(ruleNum int, qerr error) RuleError {
	return RuleError{ruleNum: ruleNum, err: qerr}
}

// Error makes RuleError comply with the error interface.
func (re RuleError) Error() string {
	pe, ok := re.err.(sexpr.ParserError)
	if !ok {
		return re.err.Error()
	}
	return fmt.Sprintf("%s in rule %d: %s", pe.ErrorType, re.ruleNum, pe.Msg)
}

// MarshalJSON makes RuleError implement the json.Marshaler interface.
func (re RuleError) MarshalJSON() ([]byte, error) {
	type errPos struct {
		Message string `json:"message"`
		Line    int    `json:"line"`
		Char    int    `json:"char"`
		AbsChar int    `json:"absChar"`
	}
	type field struct {
		Path  []string `json:"path"`
		Error errPos   `json:"error"`
	}

	var err struct {
		Error  string  `json:"error"`
		Fields []field `json:"fields"`
	}
	errMsg := re.Error()
	err.Error = errMsg
	pe, ok := re.err.(sexpr.ParserError)
	if !ok {
		return json.Marshal(err)
	}
	err.Fields = []field{
		{
			Path: []string{"rules", strconv.Itoa(re.ruleNum), "sExpr"},
			Error: errPos{
				Message: errMsg,
				Line:    pe.StartLine,
				Char:    pe.StartCharInLine,
				AbsChar: pe.StartChar,
			},
		},
	}
	return json.Marshal(err)
}
