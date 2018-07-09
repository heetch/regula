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
	r1, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
	r2, _ := rule.NewBoolRuleset(rule.New(rule.True(), rule.ReturnsBool(true)))
	l := []store.RulesetEntry{
		{Name: "a", Ruleset: r1},
		{Name: "b", Ruleset: r2},
	}

	s.AllFn = func(context.Context) ([]store.RulesetEntry, error) {
		return l, nil
	}

	h := NewHandler(s, zerolog.Nop())

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/rulesets?list", nil)
	h.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var res []store.RulesetEntry
	err := json.NewDecoder(w.Body).Decode(&res)
	require.NoError(t, err)
	require.Equal(t, res, l)
}
