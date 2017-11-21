package mock

import (
	"testing"

	rules "github.com/heetch/rules-engine"
	"github.com/stretchr/testify/require"
)

func TestMockStore(t *testing.T) {
	rss := map[string]rules.RuleSet{
		"/a/b/c": rules.RuleSet{
			rules.Rule{
				Result: rules.Result{
					Value: "abc",
					Type:  "string",
				},
				Root: &rules.NodeEq{
					Kind: "eq",
				},
			},
		},
	}

	t.Run("NewStore", func(t *testing.T) {
		s := NewStore("foobar", rss)
		require.NotNil(t, s)

		rs, err := s.Get("/a/b/c")
		require.NoError(t, err)
		require.Len(t, rs, 1)
	})
}
