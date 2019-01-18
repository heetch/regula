package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStackCanAccessRootParams(t *testing.T) {
	params := NewMockParams(map[string]interface{}{
		"foo": int64(1),
		"bar": "shoe",
		"baz": true,
	})
	stack := newStack("quux", 10.1, params)
	foo, err := stack.GetInt64("foo")
	require.NoError(t, err)
	require.Equal(t, int64(1), foo)

	bar, err := stack.GetString("bar")
	require.NoError(t, err)
	require.Equal(t, "shoe", bar)

	baz, err := stack.GetBool("baz")
	require.NoError(t, err)
	require.True(t, baz)
}

func TestStackCanAccessCurrentFrame(t *testing.T) {
	params := NewMockParams(map[string]interface{}{
		"foo": int64(1),
		"bar": "shoe",
		"baz": true,
	})
	stack := newStack("quux", 10.1, params)
	quux, err := stack.GetFloat64("quux")
	require.NoError(t, err)
	require.Equal(t, 10.1, quux)
}
