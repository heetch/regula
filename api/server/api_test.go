package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	s := new(mockStore)
	h := NewHandler(s, zerolog.Nop())

	t.Run("Root", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("List", func(t *testing.T) {
		r1, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
		r2, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
		l := []store.RulesetEntry{
			{Path: "aa", Ruleset: r1},
			{Path: "bb", Ruleset: r2},
		}

		call := func(t *testing.T, url string, code int, l []store.RulesetEntry) {
			t.Helper()

			s.ListFn = func(context.Context, string) ([]store.RulesetEntry, error) {
				return l, nil
			}
			defer func() { s.ListFn = nil }()

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", url, nil)
			h.ServeHTTP(w, r)

			require.Equal(t, code, w.Code)

			if code == http.StatusOK {
				var res []store.RulesetEntry
				err := json.NewDecoder(w.Body).Decode(&res)
				require.NoError(t, err)
				require.Equal(t, len(l), len(res))
				for i := range l {
					require.Equal(t, l[i], res[i])
				}
			}
		}

		t.Run("Root", func(t *testing.T) {
			call(t, "/rulesets/?list", http.StatusOK, l)
		})

		t.Run("WithPrefix", func(t *testing.T) {
			call(t, "/rulesets/a?list", http.StatusOK, l[:1])
		})

		t.Run("NoResultOnRoot", func(t *testing.T) {
			call(t, "/rulesets/?list", http.StatusOK, nil)
		})

		t.Run("NoResultOnPrefix", func(t *testing.T) {
			call(t, "/rulesets/someprefix?list", http.StatusNotFound, nil)
		})
	})

	t.Run("Eval", func(t *testing.T) {

		call := func(t *testing.T, url string, code int, rse *store.RulesetEntry, exp *api.Value) {
			t.Helper()

			s.OneFn = func(context.Context, string) (*store.RulesetEntry, error) {
				return rse, nil
			}
			defer func() { s.OneFn = nil }()

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
			s.OneFn = func(context.Context, string) (*store.RulesetEntry, error) {
				return nil, store.ErrNotFound
			}
			defer func() { s.OneFn = nil }()

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
}
