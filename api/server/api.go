package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
)

type rulesetService struct {
	*service

	timeout      time.Duration
	watchTimeout time.Duration
}

func (s *rulesetService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/rulesets")
	path = strings.TrimPrefix(path, "/")

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
	case "PUT":
		if path != "" {
			s.put(w, r, path)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

// list fetches all the rulesets from the store and writes them to the http response.
func (s *rulesetService) list(w http.ResponseWriter, r *http.Request, prefix string) {
	entries, err := s.store.List(r.Context(), prefix)
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	if len(entries.Entries) == 0 && prefix != "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var rl api.Rulesets

	rl.Rulesets = make([]api.Ruleset, len(entries.Entries))
	for i := range entries.Entries {
		rl.Rulesets[i] = api.Ruleset(entries.Entries[i])
	}
	rl.Revision = entries.Revision

	s.encodeJSON(w, &rl, http.StatusOK)
}

func (s *rulesetService) eval(w http.ResponseWriter, r *http.Request, path string) {
	var err error
	var e *store.RulesetEntry

	if v, ok := r.URL.Query()["version"]; ok {
		e, err = s.store.OneByVersion(r.Context(), path, v[0])
	} else {
		e, err = s.store.Latest(r.Context(), path)
	}

	if err != nil {
		if err == store.ErrNotFound {
			s.writeError(w, fmt.Errorf("the path: '%s' doesn't exist", path), http.StatusNotFound)
			return
		}
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	params := make(params)
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	v, err := e.Ruleset.Eval(params)
	if err != nil {
		if err == rule.ErrParamNotFound ||
			err == rule.ErrTypeParamMismatch ||
			err == rule.ErrNoMatch {
			s.writeError(w, err, http.StatusBadRequest)
			return
		}
		s.writeError(w, errInternal, http.StatusInternalServerError)
		return
	}

	resp := api.Value{
		Data: v.Data,
		Type: v.Type,
	}

	s.encodeJSON(w, resp, http.StatusOK)
}

// watch watches a prefix for change and returns anything newer.
func (s *rulesetService) watch(w http.ResponseWriter, r *http.Request, prefix string) {
	events, err := s.store.Watch(r.Context(), prefix, r.Header.Get("revision"))
	if err != nil {
		switch err {
		case context.DeadlineExceeded:
			w.WriteHeader(http.StatusRequestTimeout)
			return
		case store.ErrNotFound:
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			s.writeError(w, err, http.StatusInternalServerError)
			return
		}
	}

	el := api.Events{
		Events:   make([]api.Event, len(events.Events)),
		Revision: events.Revision,
	}

	for i := range events.Events {
		el.Events[i] = api.Event(events.Events[i])
	}

	s.encodeJSON(w, el, http.StatusOK)
}

// put creates a new version of a ruleset.
func (s *rulesetService) put(w http.ResponseWriter, r *http.Request, path string) {
	var rs rule.Ruleset

	err := json.NewDecoder(r.Body).Decode(&rs)
	if err != nil {
		s.writeError(w, err, http.StatusBadRequest)
		return
	}

	entry, err := s.store.Put(r.Context(), path, &rs)
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode((*api.Ruleset)(entry))
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}
}
