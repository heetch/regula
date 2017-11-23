package mock

import (
	"testing"

	"github.com/heetch/rules-engine/rule"
	"github.com/stretchr/testify/require"
)

func TestMockStore(t *testing.T) {
	rss := map[string]rule.Ruleset{
		"/a/b/c": rule.Ruleset{
			rule.New(
				rule.True(),
				rule.ReturnsStr("matched"),
			),
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
