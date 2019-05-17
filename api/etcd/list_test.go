package etcd

import (
	"context"
	"testing"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

// List returns all rulesets rulesets or not depending on the query string.
func TestList(t *testing.T) {
	rsTrue := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}
	rsFalse := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(false))}

	t.Run("OK", func(t *testing.T) {
		s, cleanup := newEtcdRulesetService(t)
		defer cleanup()

		createBoolRuleset(t, s, "c", rsTrue...)
		createBoolRuleset(t, s, "a", rsTrue...)
		createBoolRuleset(t, s, "a/1", rsTrue...)
		createBoolRuleset(t, s, "b", rsTrue...)
		createBoolRuleset(t, s, "a", rsFalse...)

		paths := []string{"a", "a/1", "b", "c"}

		rulesets, err := s.List(context.Background(), api.ListOptions{})
		require.NoError(t, err)
		require.Len(t, rulesets.Paths, len(paths))
		for i, p := range rulesets.Paths {
			require.Equal(t, paths[i], p)
		}

		require.NotEmpty(t, rulesets.Revision)
	})

	// Paging tests List with pagination.
	t.Run("Paging", func(t *testing.T) {
		s, cleanup := newEtcdRulesetService(t)
		defer cleanup()

		createBoolRuleset(t, s, "y", rsTrue...)
		createBoolRuleset(t, s, "yy", rsTrue...)
		createBoolRuleset(t, s, "y/1", rsTrue...)
		createBoolRuleset(t, s, "y/2", rsTrue...)
		createBoolRuleset(t, s, "y/3", rsTrue...)

		opt := api.ListOptions{Limit: 2}
		rulesets, err := s.List(context.Background(), opt)
		require.NoError(t, err)
		require.Len(t, rulesets.Paths, 2)
		require.Equal(t, "y", rulesets.Paths[0])
		require.Equal(t, "y/1", rulesets.Paths[1])
		require.NotEmpty(t, rulesets.Cursor)

		opt.Cursor = rulesets.Cursor
		cursor := rulesets.Cursor
		rulesets, err = s.List(context.Background(), opt)
		require.NoError(t, err)
		require.Len(t, rulesets.Paths, 2)
		require.Equal(t, "y/2", rulesets.Paths[0])
		require.Equal(t, "y/3", rulesets.Paths[1])
		require.NotEmpty(t, rulesets.Cursor)

		opt.Cursor = rulesets.Cursor
		rulesets, err = s.List(context.Background(), opt)
		require.NoError(t, err)
		require.Len(t, rulesets.Paths, 1)
		require.Equal(t, "yy", rulesets.Paths[0])
		require.Empty(t, rulesets.Cursor)

		opt.Limit = 3
		opt.Cursor = cursor
		rulesets, err = s.List(context.Background(), opt)
		require.NoError(t, err)
		require.Len(t, rulesets.Paths, 3)
		require.Equal(t, "y/2", rulesets.Paths[0])
		require.Equal(t, "y/3", rulesets.Paths[1])
		require.Equal(t, "yy", rulesets.Paths[2])
		require.Empty(t, rulesets.Cursor)

		opt.Cursor = "some cursor"
		rulesets, err = s.List(context.Background(), opt)
		require.Equal(t, api.ErrInvalidCursor, err)

		opt.Limit = -10
		opt.Cursor = ""
		rulesets, err = s.List(context.Background(), opt)
		require.NoError(t, err)
		require.Len(t, rulesets.Paths, 5)
	})
}
