package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStackCanAccessRootParams(t *testing.T) {
	params := MockParams{
		"foo": int64(1),
		"bar": "shoe",
		"baz": true,
	}
	stack := newStack("qux", 10.1, params)
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
	params := MockParams{
		"foo": int64(1),
		"bar": "shoe",
		"baz": true,
	}
	stack := newStack("qux", 10.1, params)
	qux, err := stack.GetFloat64("qux")
	require.NoError(t, err)
	require.Equal(t, 10.1, qux)
}

func TestNestedStackAccess(t *testing.T) {
	params := MockParams{
		"foo": int64(1),
		"bar": "shoe",
		"baz": true,
	}
	frame1 := newStack("qux", 10.1, params)
	frame2 := newStack("quux", "fruit", frame1)
	frame3 := newStack("corge", 99.1, frame2)

	// access frame3
	corge, err := frame3.GetFloat64("corge")
	require.NoError(t, err)
	require.Equal(t, 99.1, corge)

	// access frame2 via frame3
	quux, err := frame3.GetString("quux")
	require.NoError(t, err)
	require.Equal(t, "fruit", quux)

	// access frame1 via frame3 and frame2
	qux, err := frame3.GetFloat64("qux")
	require.NoError(t, err)
	require.Equal(t, 10.1, qux)

	// All the way back to the original Params via the stack
	baz, err := frame3.GetBool("baz")
	require.NoError(t, err)
	require.Equal(t, true, baz)

}
