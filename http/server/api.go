package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	rerrors "github.com/heetch/regula/errors"
	reghttp "github.com/heetch/regula/http"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

type rulesetAPI struct {
	rulesets api.RulesetService
	// Timeout specific for the watch requests.
	watchTimeout time.Duration
	// Context used by the Watch endpoint. Used to cancel long running
	// watch requests when the server is shutting down.
	watchCancelCtx context.Context
}

func (s *rulesetAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	r.ParseForm()

	switch r.Method {
	case "GET":
		if len(r.Form["list"]) > 0 {
			if len(path) != 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			s.list(w, r)
			return
		}
		if len(r.Form["eval"]) > 0 {
			s.eval(w, r, path)
			return
		}
		s.get(w, r, path)
		return
	case "POST":
		if len(r.Form["watch"]) > 0 {
			if len(path) != 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			ctx, cancel := context.WithTimeout(s.watchCancelCtx, s.watchTimeout)
			defer cancel()
			s.watch(w, r.WithContext(ctx))
			return
		}

		if len(path) != 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		s.create(w, r)
		return
	case "PUT":
		if path != "" {
			s.put(w, r, path)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

type createSignaturePayload struct {
	regula.Signature
	Path string `json:"path"`
}

func (s *rulesetAPI) create(w http.ResponseWriter, r *http.Request) {
	var sig createSignaturePayload

	err := json.NewDecoder(r.Body).Decode(&sig)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	err = s.rulesets.Create(r.Context(), sig.Path, &sig.Signature)
	if err != nil {
		if err == api.ErrAlreadyExists {
			writeError(w, r, err, http.StatusConflict)
			return
		}

		if err == api.ErrInvalidCursor {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		if api.IsValidationError(err) {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, &regula.Ruleset{
		Path:      sig.Path,
		Signature: &sig.Signature,
	}, http.StatusCreated)
}

func (s *rulesetAPI) get(w http.ResponseWriter, r *http.Request, path string) {
	v := r.URL.Query().Get("version")

	ruleset, err := s.rulesets.Get(r.Context(), path, v)
	if err != nil {
		if err == api.ErrRulesetNotFound {
			writeError(w, r, err, http.StatusNotFound)
			return
		}

		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, ruleset, http.StatusOK)
}

// list fetches all the rulesets paths from the store and writes them to the http response.
func (s *rulesetAPI) list(w http.ResponseWriter, r *http.Request) {
	var (
		opt api.ListOptions
		err error
	)

	if l := r.URL.Query().Get("limit"); l != "" {
		opt.Limit, err = strconv.Atoi(l)
		if err != nil {
			writeError(w, r, errors.New("invalid limit"), http.StatusBadRequest)
			return
		}
	}

	opt.Cursor = r.URL.Query().Get("cursor")

	rulesets, err := s.rulesets.List(r.Context(), opt)
	if err != nil {
		if err == api.ErrInvalidCursor {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, rulesets, http.StatusOK)
}

func (s *rulesetAPI) eval(w http.ResponseWriter, r *http.Request, path string) {
	var err error
	var res *regula.EvalResult

	params := make(params)
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	res, err = s.rulesets.Eval(r.Context(), path, r.URL.Query().Get("version"), params)
	if err != nil {
		if err == rerrors.ErrRulesetNotFound {
			writeError(w, r, fmt.Errorf("the path '%s' doesn't exist", path), http.StatusNotFound)
			return
		}

		if err == rerrors.ErrParamNotFound ||
			err == rerrors.ErrParamTypeMismatch ||
			err == rerrors.ErrNoMatch {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, res, http.StatusOK)
}

// watch is a long polling endpoint that watches a list of paths for change and returns a list of events containing all the changes
// that happened since the start of the watch.
// If the revision query param is specified, it returns anything that happened after that revision.
// If no paths are specificied, it watches any path.
// The request context can be used to limit the watch period or to cancel any running one.
func (s *rulesetAPI) watch(w http.ResponseWriter, r *http.Request) {
	var wo api.WatchOptions

	if r.ContentLength > 0 {
		// There's a non-empty body, which means that the
		// client has specified a set of paths to watch.
		err := json.NewDecoder(r.Body).Decode(&wo)
		if err != nil {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}
	}

	events, err := s.rulesets.Watch(r.Context(), wo)
	if err != nil {
		switch err {
		case context.Canceled:
			// server is probably shutting down
			// we do nothing and return a 200 to the client
		case context.DeadlineExceeded:
			// the watch request reached the deadline
			// we do nothing and return a 200 to the client
		default:
			writeError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	reghttp.EncodeJSON(w, r, events, http.StatusOK)
}

// put creates a new version of a ruleset.
func (s *rulesetAPI) put(w http.ResponseWriter, r *http.Request, path string) {
	var rules []*rule.Rule

	err := json.NewDecoder(r.Body).Decode(&rules)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	version, err := s.rulesets.Put(r.Context(), path, rules)
	if err != nil {

		if api.IsValidationError(err) {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		if err != api.ErrRulesetNotModified {
			writeError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	rs, err := s.rulesets.Get(r.Context(), path, version)
	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, rs, http.StatusOK)
}
