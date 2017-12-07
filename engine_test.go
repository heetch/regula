package rules_test

import (
	"testing"

	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	namespace string
	ruleSets  map[string]*rule.Ruleset
}

func newMockStore(namespace string, ruleSets map[string]*rule.Ruleset) *mockStore {
	return &mockStore{
		namespace: namespace,
		ruleSets:  ruleSets,
	}
}

func (s *mockStore) Get(key string) (*rule.Ruleset, error) {
	rs, ok := s.ruleSets[key]
	if !ok {
		err := rules.ErrRulesetNotFound
		return nil, err
	}

	return rs, nil
}

func TestEngine(t *testing.T) {
	m := newMockStore("/rules", map[string]*rule.Ruleset{
		"/match-string-a": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.Eq(rule.ParamStr("foo"), rule.ValueStr("bar")), rule.ReturnsStr("matched a")),
			},
		},
		"/match-string-b": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsStr("matched b")),
			},
		},
		"/type-mismatch": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Result{Type: "int", Value: "5"}),
			},
		},
		"/no-match": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.Eq(rule.ValueStr("foo"), rule.ValueStr("bar")), rule.ReturnsStr("matched d")),
			},
		},
		"/match-bool": &rule.Ruleset{
			Type: "bool",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Result{Type: "bool", Value: "true"}),
			},
		},
	})

	e := rules.NewEngine(m)
	str, err := e.GetString("/match-string-a", rule.Params{
		"foo": "bar",
	})
	require.NoError(t, err)
	require.Equal(t, "matched a", str)

	str, err = e.GetString("/match-string-b", nil)
	require.NoError(t, err)
	require.Equal(t, "matched b", str)

	b, err := e.GetBool("/match-bool", nil)
	require.NoError(t, err)
	require.True(t, b)

	_, err = e.GetString("/match-bool", nil)
	require.Equal(t, rules.ErrTypeMismatch, err)

	_, err = e.GetString("/type-mismatch", nil)
	require.Equal(t, rules.ErrTypeMismatch, err)

	_, err = e.GetString("/no-match", nil)
	require.Equal(t, rule.ErrNoMatch, err)

	_, err = e.GetString("/not-found", nil)
	require.Equal(t, rules.ErrRulesetNotFound, err)
}
