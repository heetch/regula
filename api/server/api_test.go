package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heetch/rules-engine/rule"
	"github.com/heetch/rules-engine/store"
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
}
