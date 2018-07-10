package server

import (
	"net/http"
	"strings"

	"github.com/heetch/rules-engine/api"
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
