package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
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

	reghttp.EncodeJSON(w, r, &uiError{Err: err.Error()}, code)
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

// Returns an http handler that lists all existing rulesets paths.
func (h *internalHandler) rulesetsHandler() http.Handler {

	type ruleset struct {
		Path string `json:"path"`
	}

	type response struct {
		Rulesets []ruleset `json:"rulesets"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				writeError(w, r, err, http.StatusBadRequest)
			}
			nrr := &newRulesetRequest{}
			err = json.Unmarshal(body, nrr)
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
			reghttp.EncodeJSON(w, r, nil, http.StatusOK)
		case "GET":
			var resp response
			var token string

			// run the loop at least once, no matter of the value of token
			for i := 0; i == 0 || token != ""; i++ {
				list, err := h.service.List(r.Context(), "", 100, token, true)
				if err != nil {
					writeError(w, r, err, http.StatusInternalServerError)
					return
				}

				token = list.Continue
				for _, rs := range list.Entries {
					resp.Rulesets = append(resp.Rulesets, ruleset{Path: rs.Path})
				}
			}

			reghttp.EncodeJSON(w, r, &resp, http.StatusOK)
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

type newRulesetRequest struct {
	Path      string    `json:"path"`
	Signature signature `json:"signature"`
	Rules     []rule    `json:"rules"`
}

func convertParams(input []param) (sexpr.Parameters, error) {
	parm := make(sexpr.Parameters)
	for _, p := range input {
		for k, v := range p {
			t, err := regrule.TypeFromName(v)
			if err != nil {
				return nil, err
			}
			parm[k] = t
		}
	}
	return parm, nil
}

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

//
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

type RuleError struct {
	ruleNum int
	err     error
}

func newRuleError(ruleNum int, qerr error) *RuleError {
	return &RuleError{ruleNum: ruleNum, err: err}
}

//
func (re *RuleError) Error() string {
	pe, ok := re.err.(sexpr.ParserError)
	if !ok {
		return re.err.Error()
	}
	return fmt.Sprintf("%s in rule %d: %s", pe.ErrorType, re.ruleNum, pe.Msg)
}

//
func (re *RuleError) MarshalJSON() ([]byte, error) {
	type field struct {
		Path  []string `json:"path"`
		Error struct {
			Message string `json:"message"`
			Line    int    `json:"line"`
			Char    int    `json:"char"`
			AbsChar int    `json:"absChar"`
		} `json:"error"`
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
			Path: []string{"rules", strconv.Itoa(err.ruleNum), "sExpr"},
			Error: {
				Message: errMsg,
				Line:    pe.StartLine,
				Char:    pe.StartCharInLine,
				AbsChar: StartChar,
			},
		},
	}
	return json.Marshal(err)
}
