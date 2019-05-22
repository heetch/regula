package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	path := "p/a/t/h"
	sig := &regula.Signature{ReturnType: "bool", Params: make(map[string]string)}

	t.Run("OK", func(t *testing.T) {
		rules1 := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}
		createBoolRuleset(t, s, path, rules1...)

		ruleset1, err := s.Get(context.Background(), path, "")
		require.NoError(t, err)
		require.Equal(t, path, ruleset1.Path)
		require.Len(t, ruleset1.Versions, 1)
		require.NotEmpty(t, rules1, ruleset1.Version)
		require.Equal(t, rules1, ruleset1.Rules)
		require.Equal(t, sig, ruleset1.Signature)

		// we are waiting 1 second here to avoid creating the new version in the same second as the previous one
		// ksuid gives a sorting with a one second precision
		time.Sleep(time.Second)
		rules2 := []*rule.Rule{rule.New(rule.Eq(rule.StringValue("foo"), rule.StringValue("foo")), rule.BoolValue(true))}
		createBoolRuleset(t, s, path, rules2...)

		// it should return the second version
		ruleset2, err := s.Get(context.Background(), path, "")
		require.NoError(t, err)
		require.Equal(t, path, ruleset2.Path)
		require.Len(t, ruleset2.Versions, 2)
		require.Equal(t, ruleset1.Version, ruleset2.Versions[0])
		require.Equal(t, ruleset2.Version, ruleset2.Versions[1])
		require.Equal(t, rules2, ruleset2.Rules)
		require.Equal(t, sig, ruleset2.Signature)

		// it should return the first version
		rs, err := s.Get(context.Background(), path, ruleset1.Version)
		require.NoError(t, err)
		require.Equal(t, ruleset1.Path, rs.Path)
		require.Equal(t, ruleset1.Version, rs.Version)
		require.Equal(t, ruleset1.Rules, rs.Rules)

		// it should return the second version
		rs, err = s.Get(context.Background(), path, ruleset2.Version)
		require.NoError(t, err)
		require.Equal(t, ruleset2, rs)
	})

	t.Run("Path Not found", func(t *testing.T) {
		_, err := s.Get(context.Background(), "doesntexist", "")
		require.Equal(t, err, api.ErrRulesetNotFound)
	})

	t.Run("Version Not found", func(t *testing.T) {
		rules := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}
		createBoolRuleset(t, s, path, rules...)

		_, err := s.Get(context.Background(), path, "doesntexist")
		require.Equal(t, err, api.ErrRulesetNotFound)
	})
}
