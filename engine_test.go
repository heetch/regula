package rules_test

import (
	"testing"

	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/mock"
	"github.com/heetch/rules-engine/rule"
	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	m := mock.NewStore("/rules", map[string]rule.Ruleset{
		"/a": rule.Ruleset{rule.New(rule.Eq(rule.ParamStr("foo"), rule.ValueStr("bar")), rule.ReturnsStr("matched a"))},
		"/b": rule.Ruleset{rule.New(rule.True(), rule.ReturnsStr("matched b"))},
		"/c": rule.Ruleset{rule.New(rule.True(), &rule.Result{Type: "int", Value: "5"})},
		"/d": rule.Ruleset{rule.New(rule.Eq(rule.ValueStr("foo"), rule.ValueStr("bar")), rule.ReturnsStr("matched d"))},
	})

	e := rules.NewEngine(m)
	str, err := e.GetString("/a", rule.Params{
		"foo": "bar",
	})
	require.NoError(t, err)
	require.Equal(t, "matched a", str)

	str, err = e.GetString("/b", nil)
	require.NoError(t, err)
	require.Equal(t, "matched b", str)

	_, err = e.GetString("/c", nil)
	require.Equal(t, rules.ErrTypeMismatch, err)

	_, err = e.GetString("/d", nil)
	require.Equal(t, rule.ErrNoMatch, err)

	_, err = e.GetString("/e", nil)
	require.Equal(t, rules.ErrRulesetNotFound, err)
}
