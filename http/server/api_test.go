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
	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/mock"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	s := new(mock.RulesetService)
	h := NewHandler(s, Config{
		WatchTimeout: 1 * time.Second,
	})

	r1 := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}
	r2 := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}

	t.Run("Root", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Get", func(t *testing.T) {
		rs1 := regula.Ruleset{
			Path:      "a",
			Version:   "version1",
			Rules:     r1,
			Versions:  []string{"version1"},
			Signature: &regula.Signature{ReturnType: "bool"},
		}

		rs2 := regula.Ruleset{
			Path:      "a",
			Version:   "version2",
			Rules:     r1,
			Versions:  []string{"version1", "version2"},
			Signature: &regula.Signature{ReturnType: "bool"},
		}

		tests := []struct {
			name    string
			path    string
			status  int
			ruleset *regula.Ruleset
			err     error
		}{
			{"Root", "/rulesets/a", http.StatusOK, &rs1, nil},
			{"NotFound", "/rulesets/b", http.StatusNotFound, &rs1, api.ErrRulesetNotFound},
			{"UnexpectedError", "/rulesets/a", http.StatusInternalServerError, &rs1, errors.New("unexpected error")},
			{"Specific version", "/rulesets/a?version=version2", http.StatusOK, &rs2, nil},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				uu, err := url.Parse(test.path)
				require.NoError(t, err)
				version := uu.Query().Get("version")
				s.GetFn = func(ctx context.Context, path, v string) (*regula.Ruleset, error) {
					require.Equal(t, v, version)
					return test.ruleset, err
				}
				defer func() { s.GetFn = nil }()

				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", test.path, nil)
				h.ServeHTTP(w, r)

				require.Equal(t, test.status, w.Code)

				if test.status == http.StatusOK {
					var res regula.Ruleset
					err := json.NewDecoder(w.Body).Decode(&res)
					require.NoError(t, err)
					require.Len(t, res.Versions, len(test.ruleset.Versions))
					require.Equal(t, test.ruleset.Path, res.Path)
					require.Equal(t, test.ruleset.Signature, res.Signature)
					require.Equal(t, test.ruleset.Version, res.Version)
					require.Equal(t, test.ruleset.Rules, res.Rules)
				}
			})
		}
	})

	t.Run("List", func(t *testing.T) {
		rss := api.Rulesets{
			Paths:    []string{"aa", "bb"},
			Revision: "somerev",
			Cursor:   "somecursor",
		}

		tests := []struct {
			name   string
			path   string
			status int
			rss    *api.Rulesets
			opt    api.ListOptions
			err    error
		}{
			{"OK", "/rulesets/?list", http.StatusOK, &rss, api.ListOptions{}, nil},
			{"WithLimitAndCursor", "/rulesets/a?list&limit=10&continue=abc123", http.StatusOK, &rss, api.ListOptions{Limit: 10, Cursor: "abc123"}, nil},
			{"NoResult", "/rulesets/?list", http.StatusOK, new(api.Rulesets), api.ListOptions{}, nil},
			{"InvalidCursor", "/rulesets/someprefix?list", http.StatusBadRequest, new(api.Rulesets), api.ListOptions{}, api.ErrInvalidCursor},
			{"UnexpectedError", "/rulesets/someprefix?list", http.StatusInternalServerError, new(api.Rulesets), api.ListOptions{}, errors.New("unexpected error")},
			{"InvalidLimit", "/rulesets/someprefix?list&limit=badlimit", http.StatusBadRequest, nil, api.ListOptions{}, nil},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				uu, err := url.Parse(test.path)
				require.NoError(t, err)
				limit := uu.Query().Get("limit")
				if limit == "" {
					limit = "0"
				}
				cursor := uu.Query().Get("cursor")

				s.ListFn = func(ctx context.Context, opt api.ListOptions) (*api.Rulesets, error) {
					assert.Equal(t, limit, strconv.Itoa(opt.Limit))
					assert.Equal(t, cursor, opt.Cursor)
					assert.Equal(t, test.opt, opt)
					return test.rss, err
				}
				defer func() { s.ListFn = nil }()

				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", test.path, nil)
				h.ServeHTTP(w, r)

				require.Equal(t, test.status, w.Code)

				if test.status == http.StatusOK {
					var res api.Rulesets
					err := json.NewDecoder(w.Body).Decode(&res)
					require.NoError(t, err)
					require.Equal(t, len(test.rss.Paths), len(res.Paths))
					for i := range test.rss.Paths {
						require.EqualValues(t, test.rss.Paths[i], res.Paths[i])
					}
					if len(test.rss.Paths) > 0 {
						require.Equal(t, "somecursor", res.Cursor)
					}
				}
			})
		}
	})

	t.Run("Eval", func(t *testing.T) {
		result := regula.EvalResult{Value: rule.StringValue("success"), Version: "123"}

		tests := []struct {
			name   string
			path   string
			status int
			mockFn func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
		}{
			{"OK", "/rulesets/path/to/my/ruleset?eval&str=str&nb=10&boolean=true", http.StatusOK, func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
				s, err := params.GetString("str")
				require.NoError(t, err)
				require.Equal(t, "str", s)
				i, err := params.GetInt64("nb")
				require.NoError(t, err)
				require.Equal(t, int64(10), i)
				b, err := params.GetBool("boolean")
				require.NoError(t, err)
				require.True(t, b)

				return &result, nil
			}},
			{"OK With version", "/rulesets/path/to/my/ruleset?eval&version=123&str=str&nb=10&boolean=true", http.StatusOK, func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
				s, err := params.GetString("str")
				require.NoError(t, err)
				require.Equal(t, "str", s)
				return &result, nil
			}},
			{"NOK - Ruleset not found", "/rulesets/path/to/my/ruleset?eval&foo=10", http.StatusNotFound, func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
				return nil, rerrors.ErrRulesetNotFound
			}},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				resetStore(s)

				s.EvalFn = test.mockFn

				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", test.path, nil)
				h.ServeHTTP(w, r)

				require.Equal(t, test.status, w.Code)
			})
		}

		t.Run("NOK - errors", func(t *testing.T) {
			errs := []error{
				rerrors.ErrParamNotFound,
				rerrors.ErrParamTypeMismatch,
				rerrors.ErrNoMatch,
			}

			for _, e := range errs {
				s.EvalFn = func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
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
		l := api.RulesetEvents{
			Events: []api.RulesetEvent{
				{Type: api.RulesetPutEvent, Path: "a", Rules: r1},
				{Type: api.RulesetPutEvent, Path: "b", Rules: r2},
				{Type: api.RulesetPutEvent, Path: "a", Rules: r2},
			},
			Revision: "rev",
		}

		tests := []struct {
			name   string
			path   string
			status int
			es     *api.RulesetEvents
			err    error
		}{
			{"Root", "/rulesets/?watch", http.StatusOK, &l, nil},
			{"WithPrefix", "/rulesets/a?watch", http.StatusOK, &l, nil},
			{"NotFound", "/rulesets/a?watch", http.StatusNotFound, &l, api.ErrRulesetNotFound},
			{"Timeout", "/rulesets/?watch", http.StatusOK, nil, context.DeadlineExceeded},
			{"ContextCanceled", "/rulesets/?watch", http.StatusOK, nil, context.Canceled},
		}

		for _, test := range tests {
			s.WatchFn = func(context.Context, string, string) (*api.RulesetEvents, error) {
				return test.es, test.err
			}
			defer func() { s.WatchFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", test.path, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, test.status, w.Code)

			if test.status == http.StatusOK {
				var res api.RulesetEvents
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				if test.es != nil {
					require.Equal(t, len(test.es.Events), len(res.Events))
					for i := range l.Events {
						require.Equal(t, l.Events[i], res.Events[i])
					}
				}
			}
		}

		t.Run("WithRevision", func(t *testing.T) {
			s.WatchFn = func(ctx context.Context, prefix string, revision string) (*api.RulesetEvents, error) {
				require.Equal(t, "a", prefix)
				require.Equal(t, "somerev", revision)
				return &l, nil
			}
			defer func() { s.WatchFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/rulesets/a?watch&revision=somerev", nil)
			h.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Code)

			var res api.RulesetEvents
			err := json.NewDecoder(w.Body).Decode(&res)
			require.NoError(t, err)
			require.Equal(t, len(l.Events), len(res.Events))
			for i := range l.Events {
				require.Equal(t, l.Events[i], res.Events[i])
			}
		})
	})

	t.Run("Put", func(t *testing.T) {
		tests := []struct {
			name    string
			path    string
			status  int
			version string
			err     error
		}{
			{"OK", "/rulesets/a", http.StatusOK, "version", nil},
			{"NotModified", "/rulesets/a", http.StatusOK, "version", api.ErrRulesetNotModified},
			{"EmptyPath", "/rulesets/", http.StatusNotFound, "version", nil},
			{"StoreError", "/rulesets/a", http.StatusInternalServerError, "", errors.New("some error")},
			{"Bad ruleset name", "/rulesets/a", http.StatusBadRequest, "", new(api.ValidationError)},
			{"Bad param name", "/rulesets/a", http.StatusBadRequest, "", new(api.ValidationError)},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s.PutFn = func(context.Context, string, []*rule.Rule) (string, error) {
					return test.version, test.err
				}
				defer func() { s.PutFn = nil }()

				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(r1)
				require.NoError(t, err)

				w := httptest.NewRecorder()
				r := httptest.NewRequest("PUT", test.path, &buf)
				h.ServeHTTP(w, r)

				require.Equal(t, test.status, w.Code)

				if test.status == http.StatusOK {
					var rs regula.Ruleset
					err := json.NewDecoder(w.Body).Decode(&rs)
					require.NoError(t, err)
					require.EqualValues(t, test.version, rs)
				}
			})

		}
	})

	t.Run("Create", func(t *testing.T) {
		sig := regula.Signature{ReturnType: "int64"}

		tests := []struct {
			name   string
			path   string
			status int
			err    error
		}{
			{"OK", "/rulesets/a", http.StatusCreated, nil},
			{"StoreError", "/rulesets/a", http.StatusInternalServerError, errors.New("some error")},
			{"Validation error", "/rulesets/a", http.StatusBadRequest, new(api.ValidationError)},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				resetStore(s)
				s.CreateFn = func(context.Context, string, *regula.Signature) error {
					return test.err
				}

				var buf bytes.Buffer
				err := json.NewEncoder(&buf).Encode(sig)
				require.NoError(t, err)

				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", test.path, &buf)
				h.ServeHTTP(w, r)

				require.Equal(t, test.status, w.Code)
			})
		}
	})
}

func resetStore(s *mock.RulesetService) {
	s.ListCount = 0
	s.WatchCount = 0
	s.PutCount = 0
	s.EvalCount = 0
	s.ListFn = nil
	s.WatchFn = nil
	s.PutFn = nil
	s.EvalFn = nil
}
