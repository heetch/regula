package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
)

type rulesetService struct {
	*service

	timeout      time.Duration
	watchTimeout time.Duration
}

func (s *rulesetService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/rulesets")
	path = strings.TrimPrefix(path, "/")

	if _, ok := r.URL.Query()["watch"]; ok && r.Method == "GET" {
		ctx, cancel := context.WithTimeout(r.Context(), s.watchTimeout)
		defer cancel()
		s.watch(w, r.WithContext(ctx), path)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()
	r = r.WithContext(ctx)

	switch r.Method {
	case "GET":
		if _, ok := r.URL.Query()["list"]; ok {
			s.list(w, r, path)
			return
		}
		if _, ok := r.URL.Query()["eval"]; ok {
			s.eval(w, r, path)
			return
		}
	case "PUT":
		if path != "" {
			s.put(w, r, path)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

// list fetches all the rulesets from the store and writes them to the http response.
func (s *rulesetService) list(w http.ResponseWriter, r *http.Request, prefix string) {
	entries, err := s.rulesets.List(r.Context(), prefix)
	if err != nil {
		if err == store.ErrNotFound {
			s.writeError(w, r, err, http.StatusNotFound)
			return
		}

		s.writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	var rl api.Rulesets

	rl.Rulesets = make([]api.Ruleset, len(entries.Entries))
	for i := range entries.Entries {
		rl.Rulesets[i] = api.Ruleset(entries.Entries[i])
	}
	rl.Revision = entries.Revision

	s.encodeJSON(w, r, &rl, http.StatusOK)
}

func (s *rulesetService) eval(w http.ResponseWriter, r *http.Request, path string) {
	var err error
	var res *regula.EvalResult

	params := make(params)
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	if v, ok := r.URL.Query()["version"]; ok {
		res, err = s.rulesets.EvalVersion(r.Context(), path, v[0], params)
	} else {
		res, err = s.rulesets.Eval(r.Context(), path, params)
	}

	if err != nil {
		if err == regula.ErrRulesetNotFound {
			s.writeError(w, r, fmt.Errorf("the path '%s' doesn't exist", path), http.StatusNotFound)
			return
		}

		if err == rule.ErrParamNotFound ||
			err == rule.ErrParamTypeMismatch ||
			err == rule.ErrNoMatch {
			s.writeError(w, r, err, http.StatusBadRequest)
			return
		}

		s.writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	s.encodeJSON(w, r, (*api.EvalResult)(res), http.StatusOK)
}

// watch watches a prefix for change and returns anything newer.
func (s *rulesetService) watch(w http.ResponseWriter, r *http.Request, prefix string) {
	events, err := s.rulesets.Watch(r.Context(), prefix, r.URL.Query().Get("revision"))
	if err != nil {
		switch err {
		case context.DeadlineExceeded:
			w.WriteHeader(http.StatusRequestTimeout)
			return
		case store.ErrNotFound:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			s.writeError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	el := api.Events{
		Events:   make([]api.Event, len(events.Events)),
		Revision: events.Revision,
	}

	for i := range events.Events {
		el.Events[i] = api.Event(events.Events[i])
	}

	s.encodeJSON(w, r, el, http.StatusOK)
}

// put creates a new version of a ruleset.
func (s *rulesetService) put(w http.ResponseWriter, r *http.Request, path string) {
	var rs regula.Ruleset

	err := json.NewDecoder(r.Body).Decode(&rs)
	if err != nil {
		s.writeError(w, r, err, http.StatusBadRequest)
		return
	}

	err = namingValidator(path, &rs)
	if err != nil {
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	entry, err := s.rulesets.Put(r.Context(), path, &rs)
	if err != nil && err != store.ErrNotModified {
		s.writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode((*api.Ruleset)(entry))
	if err != nil {
		s.writeError(w, r, err, http.StatusInternalServerError)
		return
	}
}

// regex used to validate ruleset name.
var rgx = regexp.MustCompile(`^[a-z]+(?:[a-z0-9-\/]?[a-z0-9])*$`)

// namingValidator validates the format of the ruleset name and
// the parameters name.
func namingValidator(path string, rs *regula.Ruleset) error {
	if ok := rgx.MatchString(path); !ok {
		return regula.ErrBadRulesetName
	}

	for _, r := range rs.Rules {
		err := paramsNameValidator(r.Expr)
		if err != nil {
			return err
		}
	}

	return nil
}

// operandsGetter is used to check if a type implements it,
// if so, we can retrieve the operands.
type operandsGetter interface {
	Operands() []rule.Expr
}

func paramsNameValidator(expr rule.Expr) error {
	if r, ok := expr.(*rule.Rule); ok {
		err := paramsNameValidator(r.Expr)
		if err != nil {
			return err
		}
	}

	if o, ok := expr.(operandsGetter); ok {
		ops := o.Operands()
		for _, op := range ops {
			err := paramsNameValidator(op)
			if err != nil {
				return err
			}
		}
	}

	if v, ok := expr.(rule.Validator); ok {
		err := v.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}
