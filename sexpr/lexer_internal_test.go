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

func TestIsRParen(t *testing.T) {
	require.True(t, isRParen(')'))
	require.False(t, isRParen(' '))
	require.False(t, isRParen('\t'))
	require.False(t, isRParen('\r'))
	require.False(t, isRParen('\n'))
	require.False(t, isRParen('-'))
	require.False(t, isRParen('a'))
	require.False(t, isRParen('"'))
	require.False(t, isRParen('('))
	require.False(t, isRParen('_'))
	require.False(t, isRParen('0'))
	require.False(t, isRParen('#'))
	require.False(t, isRParen(';'))
}

func TestIsString(t *testing.T) {
	require.True(t, isString('"'))
	require.False(t, isString(' '))
	require.False(t, isString('\t'))
	require.False(t, isString('\r'))
	require.False(t, isString('\n'))
	require.False(t, isString('-'))
	require.False(t, isString('a'))
	require.False(t, isString(')'))
	require.False(t, isString('('))
	require.False(t, isString('_'))
	require.False(t, isString('0'))
	require.False(t, isString('#'))
	require.False(t, isString(';'))
}

func TestIsNumber(t *testing.T) {
	require.True(t, isNumber('-'))
	for r := '0'; r <= '9'; r++ {
		require.True(t, isNumber(r))
	}
	require.False(t, isNumber('"'))
	require.False(t, isNumber(' '))
	require.False(t, isNumber('\t'))
	require.False(t, isNumber('\r'))
	require.False(t, isNumber('\n'))
	require.False(t, isNumber('a'))
	require.False(t, isNumber(')'))
	require.False(t, isNumber('('))
	require.False(t, isNumber('_'))
	require.False(t, isNumber('#'))
	require.False(t, isNumber(';'))
}
