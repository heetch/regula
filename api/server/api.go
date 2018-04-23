package server

import (
	"net/http"

	"github.com/heetch/rules-engine/api"
)

type rulesetService struct {
	*service
}

// list fetches all the rulesets from the store and writes them to the http response.
func (s *rulesetService) list(w http.ResponseWriter, r *http.Request) {
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
