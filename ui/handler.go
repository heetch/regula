package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/heetch/regula"
	reghttp "github.com/heetch/regula/http"
	regrule "github.com/heetch/regula/rule"
	"github.com/heetch/regula/rule/sexpr"
	"github.com/heetch/regula/store"
	"github.com/julienschmidt/httprouter"
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// NewHandler creates a http handler serving the UI application and the UI backend.
func NewHandler(service store.RulesetService, distPath string) http.Handler {
	var mux http.ServeMux

	// internal API
	mux.Handle("/i/", http.StripPrefix("/i", newInternalHandler(service)))

	// static files
	fs := http.FileServer(http.Dir(distPath))
	mux.Handle("/css/", fs)
	mux.Handle("/js/", fs)
	mux.Handle("/fonts/", fs)

	// catch all url that deleguates the routing to the front app router
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
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
	router := httprouter.New()

	// router for the internal API
	router.HandlerFunc("GET", "/rulesets/", h.handleListRequest)
	router.HandlerFunc("POST", "/rulesets/", h.handleNewRulesetRequest)
	router.Handle("PUT", "/rulesets/*path", h.handleNewRulesetVersionRequest)

	return router
}

// handleNewRulesetRequest consumes a POST to the ruleset endpoint and
// attempts to create a new Ruleset from that data.
func (h *internalHandler) handleNewRulesetRequest(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		Path      string           `json:"path"`
		Signature regula.Signature `json:"signature"`
	}

	var p payload

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	err = h.service.Create(r.Context(), p.Path, &p.Signature)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	reghttp.EncodeJSON(w, r, nil, http.StatusCreated)
}

// handleNewRulesetVersionRequest consumes a PUT to the ruleset endpoint and
// attempts to create a new Ruleset version from that data.
func (h *internalHandler) handleNewRulesetVersionRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type payload struct {
		Rules []rule `json:"rules"`
	}

	var p payload
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	path := ps.ByName("path")

	rs, err := h.service.Get(r.Context(), path, "")
	if err != nil {
		if err == store.ErrRulesetNotFound {
			writeError(w, r, err, http.StatusNotFound)
			return
		}

		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	rules, err := newRulesFromRequest(rs.Signature, p.Rules)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	_, err = h.service.Put(r.Context(), path, rules)
	if err == store.ErrRulesetNotModified {
		reghttp.EncodeJSON(w, r, nil, http.StatusNotModified)
		return
	}
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, nil, http.StatusOK)
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
		for _, rs := range list.Rulesets {
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

type rule struct {
	SExpr       string `json:"sExpr"`
	ReturnValue string `json:"returnValue"`
}

// convertParams takes a map of params and returns an equivalent sexpr.Parameters map.
func convertParams(input map[string]string) (sexpr.Parameters, error) {
	parm := make(sexpr.Parameters)
	for name, v := range input {
		if name == "" {
			return nil, errors.New("parameter has no name")
		}
		if v == "" {
			return nil, fmt.Errorf("parameter (%s) has no type", name)
		}

		t, err := regrule.TypeFromName(v)
		if err != nil {
			return nil, err
		}
		parm[name] = t
	}
	return parm, nil
}

// newRulesFromRequest takes a newRulesetVersionRequest and returns the
// equivalent []*rule.Rule.
func newRulesFromRequest(sig *regula.Signature, rawRules []rule) ([]*regrule.Rule, error) {
	var rules []*regrule.Rule
	parm, err := convertParams(sig.Params)
	if err != nil {
		return nil, err
	}

	for n, rule := range rawRules {
		p := sexpr.NewParser(bytes.NewBufferString(rule.SExpr))
		expr, err := p.Parse(parm)
		if err != nil {
			return nil, newRuleError(n+1, err)
		}
		val, err := makeValue(sig.ReturnType, rule.ReturnValue)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &regrule.Rule{
			Expr:   expr,
			Result: val,
		})

	}

	return rules, nil
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
