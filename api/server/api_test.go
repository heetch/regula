package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	s := new(mockStore)
	log := zerolog.New(ioutil.Discard)
	h := NewHandler(s, Config{
		WatchTimeout: 1 * time.Second,
		Logger:       &log,
	})

	t.Run("Root", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("List", func(t *testing.T) {
		r1, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
		r2, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
		l := store.RulesetEntries{
			Entries: []store.RulesetEntry{
				{Path: "aa", Ruleset: r1},
				{Path: "bb", Ruleset: r2},
			},
			Revision: "somerev",
		}

		call := func(t *testing.T, url string, code int, l *store.RulesetEntries) {
			t.Helper()

			s.ListFn = func(context.Context, string) (*store.RulesetEntries, error) {
				return l, nil
			}
			defer func() { s.ListFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", url, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res api.RulesetList
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Equal(t, len(l.Entries), len(res.Rulesets))
				for i := range l.Entries {
					require.EqualValues(t, l.Entries[i], res.Rulesets[i])
				}
			}
		}

		t.Run("Root", func(t *testing.T) {
			call(t, "/rulesets/?list", http.StatusOK, &l)
		})

		t.Run("WithPrefix", func(t *testing.T) {
			call(t, "/rulesets/a?list", http.StatusOK, &l)
		})

		t.Run("NoResultOnRoot", func(t *testing.T) {
			call(t, "/rulesets/?list", http.StatusOK, new(store.RulesetEntries))
		})

		t.Run("NoResultOnPrefix", func(t *testing.T) {
			call(t, "/rulesets/someprefix?list", http.StatusNotFound, new(store.RulesetEntries))
		})
	})

	t.Run("Eval", func(t *testing.T) {
		call := func(t *testing.T, url string, code int, rse *store.RulesetEntry, exp *api.Value) {
			t.Helper()

			s.LatestFn = func(context.Context, string) (*store.RulesetEntry, error) {
				return rse, nil
			}
			defer func() { s.LatestFn = nil }()

			s.OneByVersionFn = func(context.Context, string, string) (*store.RulesetEntry, error) {
				return rse, nil
			}
			defer func() { s.OneByVersionFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", url, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res api.Value
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.EqualValues(t, exp, &res)
			}
		}

		t.Run("String - OK", func(t *testing.T) {
			rs, _ := rule.NewStringRuleset(
				rule.New(
					rule.Eq(
						rule.StringParam("foo"),
						rule.StringValue("bar"),
					),
					rule.ReturnsString("success"),
				),
			)
			rse := store.RulesetEntry{
				Path:    "path/to/my/ruleset",
				Ruleset: rs,
			}

			exp := &api.Value{
				Data: "success",
				Type: "string",
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&foo=bar", http.StatusOK, &rse, exp)
			require.Equal(t, 1, s.LatestCount)
		})

		t.Run("String with version - OK", func(t *testing.T) {
			rs, _ := rule.NewStringRuleset(
				rule.New(
					rule.Eq(
						rule.StringParam("foo"),
						rule.StringValue("bar"),
					),
					rule.ReturnsString("success"),
				),
			)
			rse := store.RulesetEntry{
				Path:    "path/to/my/ruleset",
				Ruleset: rs,
			}

			exp := &api.Value{
				Data: "success",
				Type: "string",
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&version=123&foo=bar", http.StatusOK, &rse, exp)
			require.Equal(t, 1, s.OneByVersionCount)
		})

		t.Run("Bool - OK", func(t *testing.T) {
			rs, _ := rule.NewBoolRuleset(
				rule.New(
					rule.Eq(
						rule.BoolParam("foo"),
						rule.BoolValue(true),
					),
					rule.ReturnsBool(true),
				),
			)
			rse := store.RulesetEntry{
				Path:    "path/to/my/ruleset",
				Ruleset: rs,
			}

			exp := &api.Value{
				Data: "true",
				Type: "bool",
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&foo=true", http.StatusOK, &rse, exp)
		})

		t.Run("Int64 - OK", func(t *testing.T) {
			rs, _ := rule.NewInt64Ruleset(
				rule.New(
					rule.Eq(
						rule.Int64Param("foo"),
						rule.Int64Value(42),
					),
					rule.ReturnsInt64(42),
				),
			)
			rse := store.RulesetEntry{
				Path:    "path/to/my/ruleset",
				Ruleset: rs,
			}

			exp := &api.Value{
				Data: "42",
				Type: "int64",
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&foo=42", http.StatusOK, &rse, exp)
		})

		t.Run("Float64 - OK", func(t *testing.T) {
			rs, _ := rule.NewFloat64Ruleset(
				rule.New(
					rule.Eq(
						rule.Float64Param("foo"),
						rule.Float64Value(42.42),
					),
					rule.ReturnsFloat64(42.42),
				),
			)
			rse := store.RulesetEntry{
				Path:    "path/to/my/ruleset",
				Ruleset: rs,
			}

			exp := &api.Value{
				Data: "42.420000",
				Type: "float64",
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&foo=42.42", http.StatusOK, &rse, exp)
		})

		t.Run("NOK - Ruleset not found", func(t *testing.T) {
			s.LatestFn = func(context.Context, string) (*store.RulesetEntry, error) {
				return nil, store.ErrNotFound
			}
			defer func() { s.LatestFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/rulesets/path/to/my/ruleset?eval&foo=10", nil)
			h.ServeHTTP(w, r)

			exp := api.Error{
				Err: "the path: 'path/to/my/ruleset' dosn't exist",
			}

			var resp api.Error
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			require.Equal(t, http.StatusNotFound, w.Code)
			require.Equal(t, exp, resp)
		})

		t.Run("NOK - bad parameter type", func(t *testing.T) {
			rs, _ := rule.NewStringRuleset(
				rule.New(
					rule.Eq(
						rule.BoolParam("foo"),
						rule.BoolValue(true),
					),
					rule.ReturnsString("success"),
				),
			)
			rse := store.RulesetEntry{
				Path:    "path/to/my/ruleset",
				Ruleset: rs,
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&foo=bar", http.StatusBadRequest, &rse, nil)
		})

		t.Run("NOK - undefined parameter", func(t *testing.T) {
			rs, _ := rule.NewStringRuleset(
				rule.New(
					rule.Eq(
						rule.BoolParam("foo"),
						rule.BoolValue(true),
					),
					rule.ReturnsString("success"),
				),
			)
			rse := store.RulesetEntry{
				Path:    "path/to/my/ruleset",
				Ruleset: rs,
			}

			call(t, "/rulesets/path/to/my/ruleset?eval&bar=true", http.StatusBadRequest, &rse, nil)
		})
	})

	t.Run("Watch", func(t *testing.T) {
		r1, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
		r2, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
		l := store.Events{
			Events: []store.Event{
				{Type: store.PutEvent, Path: "a", Ruleset: r1},
				{Type: store.PutEvent, Path: "b", Ruleset: r2},
				{Type: store.DeleteEvent, Path: "a", Ruleset: r2},
			},
			Revision: "rev",
		}

		call := func(t *testing.T, url string, code int, es *store.Events, err error) {
			t.Helper()

			s.WatchFn = func(context.Context, string, string) (*store.Events, error) {
				return es, err
			}
			defer func() { s.WatchFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", url, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res store.Events
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Equal(t, len(l.Events), len(res.Events))
				for i := range l.Events {
					require.Equal(t, l.Events[i], res.Events[i])
				}
			}
		}

		t.Run("Root", func(t *testing.T) {
			call(t, "/rulesets/?watch", http.StatusOK, &l, nil)
		})

		t.Run("WithPrefix", func(t *testing.T) {
			call(t, "/rulesets/a?watch", http.StatusOK, &l, nil)
		})

		t.Run("WithRevision", func(t *testing.T) {
			call(t, "/rulesets/a?watch&revision=somerev", http.StatusOK, &l, nil)
		})

		t.Run("Timeout", func(t *testing.T) {
			call(t, "/rulesets/?watch", http.StatusRequestTimeout, nil, context.DeadlineExceeded)
		})
	})

	t.Run("Put", func(t *testing.T) {
		r1, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
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

		t.Run("EmptyPath", func(t *testing.T) {
			call(t, "/rulesets/", http.StatusNotFound, &e1, nil)
		})

		t.Run("StoreError", func(t *testing.T) {
			call(t, "/rulesets/a", http.StatusInternalServerError, nil, errors.New("some error"))
		})
	})
}
