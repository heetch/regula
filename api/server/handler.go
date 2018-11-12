package server

import (
	"context"
	"net/http"
	"time"

	"github.com/heetch/regula/api"
	reghttp "github.com/heetch/regula/http"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// Config contains the API configuration.
type Config struct {
	Timeout        time.Duration
	WatchTimeout   time.Duration
	WatchCancelCtx context.Context // set this to cancel watchers on demand
}

// NewHandler creates an http handler to serve the rules engine API.
func NewHandler(rsService store.RulesetService, cfg Config) http.Handler {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	if cfg.WatchTimeout == 0 {
		cfg.WatchTimeout = 30 * time.Second
	}

	if cfg.WatchCancelCtx == nil {
		cfg.WatchCancelCtx = context.Background()
	}

	rulesetsHandler := rulesetAPI{
		rulesets:     rsService,
		timeout:      cfg.Timeout,
		watchTimeout: cfg.WatchTimeout,
	}

	// router
	mux := http.NewServeMux()
	mux.HandleFunc("/rulesets/", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["watch"]; ok && r.Method == "GET" {
			rulesetsHandler.ServeHTTP(w, r.WithContext(cfg.WatchCancelCtx))
			return
		}

		rulesetsHandler.ServeHTTP(w, r)
	})

	return mux
}

// writeError writes an error to the http response in JSON format.
func writeError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// Prepare log.
	logger := reghttp.LoggerFromRequest(r).With().
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

	reghttp.EncodeJSON(w, r, &api.Error{Err: err.Error()}, code)
}
