package server

import (
	"encoding/json"
	"net/http"

	"github.com/heetch/rules-engine/store"
	"github.com/rs/zerolog"
)

type api struct {
	logger zerolog.Logger
	store  store.Store
}

// encodeJSON encodes v to w in JSON format.
func (a *api) encodeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		a.logger.Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
	}
}

// writeError write an error to the http response in JSON format.
func (a *api) writeError(w http.ResponseWriter, err error, code int) {
	// Log error.
	a.logger.Debug().Err(err).Int("code", code).Msg("http error")

	// Hide error from client if it's internal.
	if code == http.StatusInternalServerError {
		err = errInternal
	}

	a.encodeJSON(w, &httpError{Err: err.Error()}, code)
}

type httpError struct {
	Err string `json:"error"`
}

// allRulesets fetches all the rulesets from the store and writes them to the http response.
func (a *api) allRulesets(w http.ResponseWriter, r *http.Request) {
	l, err := a.store.All(r.Context())
	if err != nil {
		a.writeError(w, err, http.StatusInternalServerError)
		return
	}

	a.encodeJSON(w, l, http.StatusOK)
}
