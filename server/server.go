package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/heetch/rules-engine/store"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const (
	timeout = 5 * time.Second
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// New creates an http server to serve the rules engine API.
func New(store store.Store, logger zerolog.Logger) *http.Server {
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

	// router
	mux := httprouter.New()
	mux.HandlerFunc("GET", "/rulesets", h.allRulesets)

	// middlewares
	chain := []func(http.Handler) http.Handler{
		hlog.NewHandler(h.logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("")
		}),
		hlog.RemoteAddrHandler("ip"),
		func(http.Handler) http.Handler {
			return mux
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// setting request timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// playing the middleware chain
		var cur http.Handler
		for i := len(chain) - 1; i >= 0; i-- {
			cur = chain[i](cur)
		}

		// serving the request
		cur.ServeHTTP(w, r.WithContext(ctx))
	})
}

// encodeJSON encodes v to w in JSON format.
func (h *handler) encodeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
	}
}

func (h *handler) writeError(w http.ResponseWriter, err error, code int) {
	// Log error.
	h.logger.Debug().Err(err).Int("code", code).Msg("http error")

	// Hide error from client if it's internal.
	if code == http.StatusInternalServerError {
		err = errInternal
	}

	h.encodeJSON(w, &httpError{Err: err.Error()}, code)
}

type httpError struct {
	Err string `json:"error"`
}
