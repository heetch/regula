package ui

import (
	"errors"
	"net/http"
	"path/filepath"

	reghttp "github.com/heetch/regula/http"
	"github.com/heetch/regula/store"
)

// HTTP errors
var (
	errInternal = errors.New("internal_error")
)

// NewHandler creates a http handler serving the UI application and the UI backend.
func NewHandler(service store.RulesetService, distPath string) http.Handler {
	var mux http.ServeMux

	// internal API
	mux.Handle("/i/", http.StripPrefix("/i", newInternalHandler(service)))

	// static files
	fs := http.FileServer(http.Dir(distPath))
	mux.Handle("/css/", fs)
	mux.Handle("/js/", fs)
	mux.Handle("/fonts/", fs)

	// catch all url that deleguates the routing to the front app router
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
	})

	return &mux
}

type uiError struct {
	Err string `json:"error"`
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

	reghttp.EncodeJSON(w, r, &uiError{Err: err.Error()}, code)
}

// handler serving the UI internal API.
type internalHandler struct {
	service store.RulesetService
}

func newInternalHandler(service store.RulesetService) http.Handler {
	h := internalHandler{
		service: service,
	}
	var mux http.ServeMux

	// router for the internal API
	mux.Handle("/rulesets/", h.rulesetsHandler())

	return &mux
}

// Returns an http handler that lists all existing rulesets paths.
func (h *internalHandler) rulesetsHandler() http.Handler {
	type ruleset struct {
		Path string `json:"path"`
	}

	type response struct {
		Rulesets []ruleset `json:"rulesets"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp response
		var token string

		// run the loop at least once, no matter of the value of token
		for i := 0; i == 0 || token != ""; i++ {
			list, err := h.service.List(r.Context(), "", 100, token, true)
			if err != nil {
				writeError(w, r, err, http.StatusInternalServerError)
				return
			}

			token = list.Continue
			for _, rs := range list.Entries {
				resp.Rulesets = append(resp.Rulesets, ruleset{Path: rs.Path})
			}
		}

		reghttp.EncodeJSON(w, r, &resp, http.StatusOK)
	})
}
