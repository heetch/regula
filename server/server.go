package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/heetch/rules-engine/store"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
)

const (
	timeout = 5 * time.Second
)

// NewServer creates an http server to serve the rules engine API.
func NewServer(store store.Store, logger zerolog.Logger) *http.Server {
	return &http.Server{
		Handler: newHandler(store, logger),
	}
}

type handler struct {
	logger zerolog.Logger
	store  store.Store
}

func newHandler(store store.Store, logger zerolog.Logger) http.Handler {
	h := handler{
		store:  store,
		logger: logger,
	}

	mux := httprouter.New()
	mux.HandlerFunc("GET", "/rulesets", h.allRulesets)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

// encodeJSON encodes v to w in JSON format.
func (h *handler) encodeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
	}
}

func (h *handler) allRulesets(w http.ResponseWriter, r *http.Request) {
	l, err := h.store.All(r.Context())
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}

	h.encodeJSON(w, l, http.StatusOK)
}
