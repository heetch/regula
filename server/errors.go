package server

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

func (h *handler) writeError(w http.ResponseWriter, err error, code int) {
	// Log error.
	h.logger.Debug().Err(err).Int("code", code).Msg("http error")

	// Hide error from client if it's internal.
	if code == http.StatusInternalServerError {
		err = errInternal
	}

	w.WriteHeader(code)
	if err == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&httpError{Err: err.Error()})
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to encode http error")
	}
}

type httpError struct {
	Err string `json:"error"`
}
