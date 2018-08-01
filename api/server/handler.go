package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// Config contains the API configuration.
type Config struct {
	Logger       *zerolog.Logger
	WatchTimeout time.Duration
}

// NewHandler creates an http handler to serve the rules engine API.
func NewHandler(ctx context.Context, rsService store.RulesetService, cfg Config) http.Handler {
	s := service{
		rulesets: rsService,
	}

	var logger zerolog.Logger

	if cfg.Logger != nil {
		logger = *cfg.Logger
	} else {
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	rs := rulesetService{
		service:      &s,
		watchTimeout: 60 * time.Second,
		timeout:      5 * time.Second,
	}

	// router
	mux := http.NewServeMux()
	mux.Handle("/rulesets/", &rs)

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
			return mux
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// playing the middleware chain
		var cur http.Handler
		for i := len(chain) - 1; i >= 0; i-- {
			cur = chain[i](cur)
		}

		// serving the request
		cur.ServeHTTP(w, r.WithContext(ctx))
	})
}

type service struct {
	rulesets store.RulesetService
}

// encodeJSON encodes v to w in JSON format.
func (s *service) encodeJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		loggerFromRequest(r).Error().Err(err).Interface("value", v).Msg("failed to encode value to http response")
	}
}

func loggerFromRequest(r *http.Request) *zerolog.Logger {
	logger := hlog.FromRequest(r).With().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Logger()
	return &logger
}

// writeError writes an error to the http response in JSON format.
func (s *service) writeError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// Prepare log.
	logger := loggerFromRequest(r).With().
		Err(err).
		Int("status", code).
		Logger()

	// Hide error from client if it's internal.
	if code == http.StatusInternalServerError {
		logger.Error().Msg("unexpected http error")
		err = errInternal
	} else {
		logger.Debug().Msg("http error")
	}

	s.encodeJSON(w, r, &api.Error{Err: err.Error()}, code)
}
