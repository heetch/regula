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

	switch r.Method {
	case "GET":
		if _, ok := r.URL.Query()["list"]; ok {
			s.list(w, r, path)
			return
		}
	}
}

// list fetches all the rulesets from the store and writes them to the http response.
func (s *rulesetService) list(w http.ResponseWriter, r *http.Request, path string) {
	l, err := s.store.All(r.Context())
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	rl := make([]api.Ruleset, len(l))
	for i := range l {
		rl[i] = api.Ruleset(l[i])
	}

	s.encodeJSON(w, rl, http.StatusOK)
}
