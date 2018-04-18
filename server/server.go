package server

import (
	"context"
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

func newHandler(store store.Store, logger zerolog.Logger) http.Handler {
	a := api{
		store:  store,
		logger: logger,
	}

	// router
	mux := httprouter.New()
	mux.HandlerFunc("GET", "/rulesets", a.allRulesets)

	// middlewares
	chain := []func(http.Handler) http.Handler{
		hlog.NewHandler(a.logger),
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
