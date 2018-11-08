// Package http provides helpers for writing http servers and clients.
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// NewHandler creates an http handler with logging capabilities.
// Every incoming request is logged automatically using the given logger, displaying various
// informations about the request (status, method, etc.)
// The provided logger is injected in the request context and can be retrieved using the LoggerFromRequest function.
func NewHandler(logger zerolog.Logger, handler http.Handler) http.Handler {
	// middlewares
	chain := []func(http.Handler) http.Handler{
		hlog.NewHandler(logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("request received")
		}),
		hlog.RemoteAddrHandler("ip"),
		hlog.UserAgentHandler("user_agent"),
		hlog.RefererHandler("referer"),
		func(http.Handler) http.Handler {
			return handler
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// playing the middleware chain
		var cur http.Handler
		for i := len(chain) - 1; i >= 0; i-- {
			cur = chain[i](cur)
		}

		// serving the request
		cur.ServeHTTP(w, r)
	})
}

// EncodeJSON encodes v to w in JSON format and fills the Content-Type header.
func EncodeJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		LoggerFromRequest(r).Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
	}
}

// LoggerFromRequest extracts the logger from the given http request.
func LoggerFromRequest(r *http.Request) *zerolog.Logger {
	logger := hlog.FromRequest(r).With().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Logger()
	return &logger
}
