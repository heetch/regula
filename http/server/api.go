package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	rerrors "github.com/heetch/regula/errors"
	reghttp "github.com/heetch/regula/http"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

type rulesetAPI struct {
	rulesets     api.RulesetService
	timeout      time.Duration
	watchTimeout time.Duration
}

func (s *rulesetAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/rulesets/")

	if _, ok := r.URL.Query()["watch"]; ok && r.Method == "GET" {
		ctx, cancel := context.WithTimeout(r.Context(), s.watchTimeout)
		defer cancel()
		s.watch(w, r.WithContext(ctx), path)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()
	r = r.WithContext(ctx)

	switch r.Method {
	case "GET":
		if _, ok := r.URL.Query()["list"]; ok {
			s.list(w, r, path)
			return
		}
		if _, ok := r.URL.Query()["eval"]; ok {
			s.eval(w, r, path)
			return
		}
		s.get(w, r, path)
		return
	case "POST":
		s.create(w, r, path)
		return
	case "PUT":
		if path != "" {
			s.put(w, r, path)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s *rulesetAPI) create(w http.ResponseWriter, r *http.Request, path string) {
	var sig regula.Signature

	err := json.NewDecoder(r.Body).Decode(&sig)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	err = s.rulesets.Create(r.Context(), path, &sig)
	if err != nil {
		if err == api.ErrAlreadyExists {
			writeError(w, r, err, http.StatusConflict)
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
		Path:      path,
		Signature: &sig,
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

// list fetches all the rulesets from the store and writes them to the http response if
// the paths parameter is not given otherwise it fetches the rulesets paths only.
func (s *rulesetAPI) list(w http.ResponseWriter, r *http.Request, prefix string) {
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
	_, opt.PathsOnly = r.URL.Query()["paths"]
	_, opt.AllVersions = r.URL.Query()["versions"]
	if opt.PathsOnly && opt.AllVersions {
		writeError(w, r, errors.New("'paths' and 'versions' parameters can't be given in the same query"), http.StatusBadRequest)
		return
	}

	rulesets, err := s.rulesets.List(r.Context(), prefix, &opt)
	if err != nil {
		if err == api.ErrRulesetNotFound {
			writeError(w, r, err, http.StatusNotFound)
			return
		}

		if err == api.ErrInvalidCursor {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, (*api.Rulesets)(rulesets), http.StatusOK)
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

// watch watches a prefix for change and returns anything newer.
func (s *rulesetAPI) watch(w http.ResponseWriter, r *http.Request, prefix string) {
	events, err := s.rulesets.Watch(r.Context(), prefix, r.URL.Query().Get("revision"))
	if err != nil {
		switch err {
		case context.Canceled, context.DeadlineExceeded:
			// we do nothing
		case api.ErrRulesetNotFound:
			w.WriteHeader(http.StatusNotFound)
			return
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

	ruleset, err := s.rulesets.Put(r.Context(), path, rules)
	if err != nil && err != api.ErrRulesetNotModified {
		if api.IsValidationError(err) {
			writeError(w, r, err, http.StatusBadRequest)
			return
		}

		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	reghttp.EncodeJSON(w, r, ruleset, http.StatusOK)
}
