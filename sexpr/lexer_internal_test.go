package sexpr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsWhitespace(t *testing.T) {
	require.True(t, isWhitespace(' '))
	require.True(t, isWhitespace('\t'))
	require.True(t, isWhitespace('\r'))
	require.True(t, isWhitespace('\n'))
	require.False(t, isWhitespace('-'))
	require.False(t, isWhitespace('a'))
	require.False(t, isWhitespace('"'))
	require.False(t, isWhitespace('('))
	require.False(t, isWhitespace(')'))
	require.False(t, isWhitespace('_'))
	require.False(t, isWhitespace('0'))
	require.False(t, isWhitespace('#'))
	require.False(t, isWhitespace(';'))
}

func TestIsLParen(t *testing.T) {
	require.True(t, isLParen('('))
	require.False(t, isLParen(' '))
	require.False(t, isLParen('\t'))
	require.False(t, isLParen('\r'))
	require.False(t, isLParen('\n'))
	require.False(t, isLParen('-'))
	require.False(t, isLParen('a'))
	require.False(t, isLParen('"'))
	require.False(t, isLParen(')'))
	require.False(t, isLParen('_'))
	require.False(t, isLParen('0'))
	require.False(t, isLParen('#'))
	require.False(t, isLParen(';'))

}
