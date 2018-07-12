package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/heetch/rules-engine/api"
	"github.com/heetch/rules-engine/rule"
	"github.com/heetch/rules-engine/store"
)

type rulesetService struct {
	*service
}

func (s *rulesetService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/rulesets")
	path = strings.TrimPrefix(path, "/")

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
	}

	w.WriteHeader(http.StatusNotFound)
}

// list fetches all the rulesets from the store and writes them to the http response.
func (s *rulesetService) list(w http.ResponseWriter, r *http.Request, prefix string) {
	l, err := s.store.List(r.Context(), prefix)
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	if len(l) == 0 && prefix != "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rl := make([]api.Ruleset, len(l))
	for i := range l {
		rl[i] = api.Ruleset(l[i])
	}

	s.encodeJSON(w, rl, http.StatusOK)
}

func (s *rulesetService) eval(w http.ResponseWriter, r *http.Request, path string) {
	e, err := s.store.One(r.Context(), path)
	if err != nil {
		if err == store.ErrNotFound {
			s.writeError(w, fmt.Errorf("the path: '%s' dosn't exist", path), http.StatusNotFound)
			return
		}
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	params := make(params)
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	v, err := e.Ruleset.Eval(params)
	if err != nil {
		if err == rule.ErrParamNotFound ||
			err == rule.ErrTypeParamMismatch ||
			err == rule.ErrNoMatch {
			s.writeError(w, err, http.StatusBadRequest)
			return
		}
		s.writeError(w, errInternal, http.StatusInternalServerError)
		return
	}

	resp := api.Response{
		Data: v.Data,
		Type: v.Type,
	}

	s.encodeJSON(w, resp, http.StatusOK)
}
