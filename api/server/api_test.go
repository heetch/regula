package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/mock"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	s := new(mock.RulesetService)
	h := NewHandler(s, Config{
		WatchTimeout: 1 * time.Second,
	})

	t.Run("Root", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Get", func(t *testing.T) {
		r1, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
		e1 := store.RulesetEntry{
			Path:      "a",
			Version:   "version",
			Ruleset:   r1,
			Versions:  []string{"version"},
			Signature: store.NewSignature(r1),
		}

		e2 := store.RulesetEntry{
			Path:      "a",
			Version:   "version2",
			Ruleset:   r1,
			Versions:  []string{"version1", "version2"},
			Signature: store.NewSignature(r1),
		}

		call := func(t *testing.T, u string, code int, e *store.RulesetEntry, err error) {
			t.Helper()

			uu, uerr := url.Parse(u)
			require.NoError(t, uerr)
			version := uu.Query().Get("version")
			s.GetFn = func(ctx context.Context, path, v string) (*store.RulesetEntry, error) {
				require.Equal(t, v, version)
				return e, err
			}
			defer func() { s.GetFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", u, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res api.Ruleset
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Len(t, res.Versions, len(e.Versions))
				require.Equal(t, e.Path, res.Path)
				require.Equal(t, e.Signature, res.Signature)
				require.Equal(t, e.Version, res.Version)
				require.Equal(t, e.Ruleset, res.Ruleset)
			}
		}

		t.Run("Root", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusOK, &e1, nil)
		})

		t.Run("NotFound", func(t *testing.T) {
			call(t, "/rulesets/b", http.StatusNotFound, &e1, store.ErrNotFound)
		})

		t.Run("UnexpectedError", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusInternalServerError, &e1, errors.New("unexpected error"))
		})

		t.Run("Specific version", func(t *testing.T) {
			call(t, "/rulesets/a?version=version2", http.StatusOK, &e2, nil)
		})
	})

	t.Run("List", func(t *testing.T) {
		r1, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
		r2, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
		l := store.RulesetEntries{
			Entries: []store.RulesetEntry{
				{Path: "aa", Ruleset: r1},
				{Path: "bb", Ruleset: r2},
			},
			Revision: "somerev",
			Continue: "sometoken",
		}

		call := func(t *testing.T, u string, code int, l *store.RulesetEntries, err error) {
			t.Helper()

			uu, uerr := url.Parse(u)
			require.NoError(t, uerr)
			limit := uu.Query().Get("limit")
			if limit == "" {
				limit = "0"
			}
			token := uu.Query().Get("continue")

			s.ListFn = func(ctx context.Context, prefix string, lm int, tk string, pathsOnly bool) (*store.RulesetEntries, error) {
				assert.Equal(t, limit, strconv.Itoa(lm))
				assert.Equal(t, token, tk)
				return l, err
			}
			defer func() { s.ListFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", u, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res api.Rulesets
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Equal(t, len(l.Entries), len(res.Rulesets))
				for i := range l.Entries {
					require.EqualValues(t, l.Entries[i], res.Rulesets[i])
				}
				if len(l.Entries) > 0 {
					require.Equal(t, "sometoken", res.Continue)
				}
			}
		}

		t.Run("Root", func(t *testing.T) {
			call(t, "/rulesets/?list", http.StatusOK, &l, nil)
		})

		t.Run("WithPrefix", func(t *testing.T) {
			call(t, "/rulesets/a?list", http.StatusOK, &l, nil)
		})

		t.Run("WithLimitAndContinue", func(t *testing.T) {
			call(t, "/rulesets/a?list&limit=10&continue=abc123", http.StatusOK, &l, nil)
		})

		t.Run("NoResultOnRoot", func(t *testing.T) {
			call(t, "/rulesets/?list", http.StatusOK, new(store.RulesetEntries), nil)
		})

		t.Run("NoResultOnPrefix", func(t *testing.T) {
			call(t, "/rulesets/someprefix?list", http.StatusNotFound, new(store.RulesetEntries), store.ErrNotFound)
		})

		t.Run("InvalidToken", func(t *testing.T) {
			call(t, "/rulesets/someprefix?list", http.StatusBadRequest, new(store.RulesetEntries), store.ErrInvalidContinueToken)
		})

		t.Run("UnexpectedError", func(t *testing.T) {
			call(t, "/rulesets/someprefix?list", http.StatusInternalServerError, new(store.RulesetEntries), errors.New("unexpected error"))
		})

		t.Run("InvalidLimit", func(t *testing.T) {
			call(t, "/rulesets/someprefix?list&limit=badlimit", http.StatusBadRequest, nil, nil)
		})
	})

	t.Run("ListPaths", func(t *testing.T) {
		l := store.RulesetEntries{
			Entries: []store.RulesetEntry{
				{Path: "aa"},
				{Path: "bb"},
			},
			Revision: "somerev",
			Continue: "sometoken",
		}

		call := func(t *testing.T, u string, code int, l *store.RulesetEntries, err error) {
			t.Helper()

			uu, uerr := url.Parse(u)
			require.NoError(t, uerr)
			limit := uu.Query().Get("limit")
			if limit == "" {
				limit = "0"
			}
			token := uu.Query().Get("continue")

			s.ListFn = func(ctx context.Context, prefix string, lm int, tk string, pathsOnly bool) (*store.RulesetEntries, error) {
				assert.Equal(t, limit, strconv.Itoa(lm))
				assert.Equal(t, token, tk)
				return l, err
			}
			defer func() { s.ListFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", u, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res api.Rulesets
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Equal(t, len(l.Entries), len(res.Rulesets))
				for i := range l.Entries {
					require.Equal(t, l.Entries[i].Path, res.Rulesets[i].Path)
					require.Zero(t, l.Entries[i].Ruleset)
					require.Zero(t, l.Entries[i].Version)
				}
				if len(l.Entries) > 0 {
					require.Equal(t, "sometoken", res.Continue)
				}
			}
		}

		t.Run("Root", func(t *testing.T) {
			call(t, "/rulesets/?list&paths", http.StatusOK, &l, nil)
		})

		t.Run("WithPrefix", func(t *testing.T) {
			call(t, "/rulesets/a?list&paths", http.StatusOK, &l, nil)
		})

		t.Run("WithLimitAndContinue", func(t *testing.T) {
			call(t, "/rulesets/a?list&paths&limit=10&continue=abc123", http.StatusOK, &l, nil)
		})
	})

	t.Run("Eval", func(t *testing.T) {
		call := func(t *testing.T, url string, code int, result *api.EvalResult, testParamsFn func(params rule.Params)) {
			t.Helper()
			resetStore(s)

			s.EvalFn = func(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error) {
				testParamsFn(params)
				return (*regula.EvalResult)(result), nil
			}

			s.EvalVersionFn = func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
				return (*regula.EvalResult)(result), nil
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", url, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res api.EvalResult
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.EqualValues(t, result, &res)
			}
		}

		t.Run("OK", func(t *testing.T) {
			exp := api.EvalResult{
				Value: rule.StringValue("success"),
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&str=str&nb=10&boolean=true", http.StatusOK, &exp, func(params rule.Params) {
				s, err := params.GetString("str")
				require.NoError(t, err)
				require.Equal(t, "str", s)
				i, err := params.GetInt64("nb")
				require.NoError(t, err)
				require.Equal(t, int64(10), i)
				b, err := params.GetBool("boolean")
				require.NoError(t, err)
				require.True(t, b)
			})
			require.Equal(t, 1, s.EvalCount)
		})

		t.Run("OK With version", func(t *testing.T) {
			exp := api.EvalResult{
				Value:   rule.StringValue("success"),
				Version: "123",
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&version=123&str=str&nb=10&boolean=true", http.StatusOK, &exp, func(params rule.Params) {
				s, err := params.GetString("str")
				require.NoError(t, err)
				require.Equal(t, "str", s)
			})
			require.Equal(t, 1, s.EvalVersionCount)
		})

		t.Run("NOK - Ruleset not found", func(t *testing.T) {
			s.EvalFn = func(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error) {
				return nil, regula.ErrRulesetNotFound
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/rulesets/path/to/my/ruleset?eval&foo=10", nil)
			h.ServeHTTP(w, r)

			exp := api.Error{
				Err: "the path 'path/to/my/ruleset' doesn't exist",
			}

			var resp api.Error
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			require.Equal(t, http.StatusNotFound, w.Code)
			require.Equal(t, exp, resp)
		})

		t.Run("NOK - errors", func(t *testing.T) {
			errs := []error{
				rule.ErrParamNotFound,
				rule.ErrParamTypeMismatch,
				rule.ErrNoMatch,
			}

			for _, e := range errs {
				s.EvalFn = func(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error) {
					return nil, e
				}

				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/rulesets/path/to/my/ruleset?eval&foo=10", nil)
				h.ServeHTTP(w, r)

				require.Equal(t, http.StatusBadRequest, w.Code)
			}
		})
	})

	t.Run("Watch", func(t *testing.T) {
		r1, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
		r2, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
		l := store.RulesetEvents{
			Events: []store.RulesetEvent{
				{Type: store.RulesetPutEvent, Path: "a", Ruleset: r1},
				{Type: store.RulesetPutEvent, Path: "b", Ruleset: r2},
				{Type: store.RulesetPutEvent, Path: "a", Ruleset: r2},
			},
			Revision: "rev",
		}

		call := func(t *testing.T, url string, code int, es *store.RulesetEvents, err error) {
			t.Helper()

			s.WatchFn = func(context.Context, string, string) (*store.RulesetEvents, error) {
				return es, err
			}
			defer func() { s.WatchFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", url, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res store.RulesetEvents
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				if es != nil {
					require.Equal(t, len(es.Events), len(res.Events))
					for i := range l.Events {
						require.Equal(t, l.Events[i], res.Events[i])
					}
				}
			}
		}

		t.Run("Root", func(t *testing.T) {
			call(t, "/rulesets/?watch", http.StatusOK, &l, nil)
		})

		t.Run("WithPrefix", func(t *testing.T) {
			call(t, "/rulesets/a?watch", http.StatusOK, &l, nil)
		})

		t.Run("NotFound", func(t *testing.T) {
			call(t, "/rulesets/a?watch", http.StatusNotFound, &l, store.ErrNotFound)
		})

		t.Run("WithRevision", func(t *testing.T) {
			t.Helper()

			s.WatchFn = func(ctx context.Context, prefix string, revision string) (*store.RulesetEvents, error) {
				require.Equal(t, "a", prefix)
				require.Equal(t, "somerev", revision)
				return &l, nil
			}
			defer func() { s.WatchFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/rulesets/a?watch&revision=somerev", nil)
			h.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Code)

			var res store.RulesetEvents
			err := json.NewDecoder(w.Body).Decode(&res)
			require.NoError(t, err)
			require.Equal(t, len(l.Events), len(res.Events))
			for i := range l.Events {
				require.Equal(t, l.Events[i], res.Events[i])
			}
		})

		t.Run("Timeout", func(t *testing.T) {
			call(t, "/rulesets/?watch", http.StatusOK, nil, context.DeadlineExceeded)
		})

		t.Run("ContextCanceled", func(t *testing.T) {
			call(t, "/rulesets/?watch", http.StatusOK, nil, context.Canceled)
		})
	})

	t.Run("Put", func(t *testing.T) {
		r1, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
		e1 := store.RulesetEntry{
			Path:    "a",
			Version: "version",
			Ruleset: r1,
		}

		call := func(t *testing.T, url string, code int, e *store.RulesetEntry, putErr error) {
			t.Helper()

			s.PutFn = func(context.Context, string) (*store.RulesetEntry, error) {
				return e, putErr
			}
			defer func() { s.PutFn = nil }()

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(r1)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("PUT", url, &buf)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var rs api.Ruleset
				err := json.NewDecoder(w.Body).Decode(&rs)
				require.NoError(t, err)
				require.EqualValues(t, *e, rs)
			}
		}

		t.Run("OK", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusOK, &e1, nil)
		})

		t.Run("NotModified", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusOK, &e1, store.ErrNotModified)
		})

		t.Run("EmptyPath", func(t *testing.T) {
			call(t, "/rulesets/", http.StatusNotFound, &e1, nil)
		})

		t.Run("StoreError", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusInternalServerError, nil, errors.New("some error"))
		})

		t.Run("Bad ruleset name", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusBadRequest, nil, new(store.ValidationError))
		})

		t.Run("Bad param name", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusBadRequest, nil, new(store.ValidationError))
		})
	})
}

func resetStore(s *mock.RulesetService) {
	s.ListCount = 0
	s.WatchCount = 0
	s.PutCount = 0
	s.EvalCount = 0
	s.EvalVersionCount = 0
	s.ListFn = nil
	s.WatchFn = nil
	s.PutFn = nil
	s.EvalFn = nil
	s.EvalVersionFn = nil
}
