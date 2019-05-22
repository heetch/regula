package etcd

import (
	"context"
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	sig := &regula.Signature{ReturnType: "bool", Params: map[string]string{"id": "string"}}
	require.NoError(t, s.Create(context.Background(), "a", sig))

	r := rule.New(
		rule.Eq(
			rule.StringParam("id"),
			rule.StringValue("123"),
		),
		rule.BoolValue(true),
	)

	ruleset := createRuleset(t, s, "a", r)
	version := ruleset.Version

	t.Run("SpecificVersion", func(t *testing.T) {
		res, err := s.Eval(context.Background(), "a", version, regula.Params{
			"id": "123",
		})
		require.NoError(t, err)
		require.Equal(t, version, res.Version)
		require.Equal(t, rule.BoolValue(true), res.Value)
	})

	t.Run("Latest", func(t *testing.T) {
		res, err := s.Eval(context.Background(), "a", "", regula.Params{
			"id": "123",
		})
		require.NoError(t, err)
		require.Equal(t, version, res.Version)
		require.Equal(t, rule.BoolValue(true), res.Value)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := s.Eval(context.Background(), "b", version, regula.Params{
			"id": "123",
		})
		require.Equal(t, errors.ErrRulesetNotFound, err)
	})

	t.Run("BadVersion", func(t *testing.T) {
		_, err := s.Eval(context.Background(), "a", "someversion", regula.Params{
			"id": "123",
		})
		require.Equal(t, errors.ErrRulesetNotFound, err)
	})
}
