package sexpr

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsWhitespace(t *testing.T) {
	require.True(t, isWhitespace(' '))
	require.True(t, isWhitespace('\v'))
	require.True(t, isWhitespace('\f'))
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

func TestIsBool(t *testing.T) {
	require.True(t, isBool('#'))
	require.False(t, isBool('"'))
	require.False(t, isBool(' '))
	require.False(t, isBool('\t'))
	require.False(t, isBool('\r'))
	require.False(t, isBool('\n'))
	require.False(t, isBool('-'))
	require.False(t, isBool('a'))
	require.False(t, isBool(')'))
	require.False(t, isBool('('))
	require.False(t, isBool('_'))
	require.False(t, isBool('0'))
	require.False(t, isBool(';'))
}

func TestIsComment(t *testing.T) {
	require.True(t, isComment(';'))
	require.False(t, isComment('#'))
	require.False(t, isComment('"'))
	require.False(t, isComment(' '))
	require.False(t, isComment('\t'))
	require.False(t, isComment('\r'))
	require.False(t, isComment('\n'))
	require.False(t, isComment('-'))
	require.False(t, isComment('a'))
	require.False(t, isComment(')'))
	require.False(t, isComment('('))
	require.False(t, isComment('_'))
	require.False(t, isComment('0'))
}

func TestIsSymbol(t *testing.T) {
	require.True(t, isSymbol('a'))
	require.True(t, isSymbol('Z'))
	require.True(t, isSymbol('!'))
	require.True(t, isSymbol('+'))
	require.True(t, isSymbol('_'))

	require.False(t, isSymbol(';'))
	require.False(t, isSymbol('#'))
	require.False(t, isSymbol('"'))
	require.False(t, isSymbol(' '))
	require.False(t, isSymbol('\t'))
	require.False(t, isSymbol('\r'))
	require.False(t, isSymbol('\n'))

	// '-' is a special case because it can also denote a number -
	// we'll have to handle this in the parser
	require.False(t, isSymbol('-'))

	require.False(t, isSymbol(')'))
	require.False(t, isSymbol('('))
	require.False(t, isSymbol('0'))
}

// NewScanner wraps an io.Reader
func TestNewScanner(t *testing.T) {
	expected := "(+ 1 1)"
	b := bytes.NewBufferString(expected)
	s := NewScanner(b)
	content, err := s.r.ReadString('\n')
	require.Error(t, err)
	require.Equal(t, io.EOF, err)
	require.Equal(t, expected, content)
}

func assertScanned(t *testing.T, input, output string, token Token, byteCount, charCount, lineCount, lineCharCount int) {
	t.Run(fmt.Sprintf("Scan %s 0x%x", input, input), func(t *testing.T) {
		b := bytes.NewBufferString(input)
		s := NewScanner(b)
		tok, lit, err := s.Scan()
		require.NoError(t, err)
		require.Equal(t, token, tok)
		require.Equal(t, output, lit)
		require.Equalf(t, byteCount, s.byteCount, "byteCount")
		require.Equalf(t, charCount, s.charCount, "charCount")
		require.Equalf(t, lineCount, s.lineCount, "lineCount")
		require.Equalf(t, lineCharCount, s.lineCharCount, "lineCharCount")
	})
}

func assertScanFailed(t *testing.T, input, message string) {
	t.Run(fmt.Sprintf("Scan should fail %s 0x%x", input, input), func(t *testing.T) {
		b := bytes.NewBufferString(input)
		s := NewScanner(b)
		_, _, err := s.Scan()
		require.EqualError(t, err, message)
	})

}

func TestScannerScan(t *testing.T) {
	// Test L Parenthesis
	assertScanned(t, "(", "(", LPAREN, 1, 1, 1, 1)
	// Test R Parenthesis
	assertScanned(t, ")", ")", RPAREN, 1, 1, 1, 1)
	// Test white-space
	assertScanned(t, " ", " ", WHITESPACE, 1, 1, 1, 1)
	assertScanned(t, "\t", "\t", WHITESPACE, 1, 1, 1, 1)
	assertScanned(t, "\r", "\r", WHITESPACE, 1, 1, 1, 1)
	assertScanned(t, "\n", "\n", WHITESPACE, 1, 1, 2, 0)
	assertScanned(t, "\v", "\v", WHITESPACE, 1, 1, 1, 1)
	assertScanned(t, "\f", "\f", WHITESPACE, 1, 1, 1, 1)
	// Test contiguous white-space:
	// - terminated by EOF
	assertScanned(t, "  ", "  ", WHITESPACE, 2, 2, 1, 2)
	// - terminated by non white-space character.
	assertScanned(t, "  (", "  ", WHITESPACE, 2, 2, 1, 2)
	// Test string:
	// - the empty string
	assertScanned(t, `""`, "", STRING, 2, 2, 1, 2)
	// - the happy case
	assertScanned(t, `"foo"`, "foo", STRING, 5, 5, 1, 5)
	// - an unterminated sad case
	assertScanFailed(t, `"foo`, "Error:1,4: unterminated string constant")
	// - happy case with escaped double quote
	assertScanned(t, `"foo\""`, `foo"`, STRING, 7, 7, 1, 7)
	// - sad case with escaped terminator
	assertScanFailed(t, `"foo\"`, "Error:1,6: unterminated string constant")
	// Test number
	// - Single digit integer, EOF terminated
	assertScanned(t, "1", "1", NUMBER, 1, 1, 1, 1)
	// - Single digit integer, terminated by non-numeric character
	assertScanned(t, "1)", "1", NUMBER, 1, 1, 1, 1)
	// - Multi-digit integer, EOF terminated
	assertScanned(t, "998989", "998989", NUMBER, 6, 6, 1, 6)
	// - Negative multi-digit integer, EOF terminated
	assertScanned(t, "-100", "-100", NUMBER, 4, 4, 1, 4)
	// - Floating point number, EOF terminated
	assertScanned(t, "2.4", "2.4", NUMBER, 3, 3, 1, 3)
	// - long negative float, terminated by non-numeric character
	assertScanned(t, "-123.45456 ", "-123.45456", NUMBER, 10, 10, 1, 10)
	// - special case: a "-" without a number following it (as per the minus operator)
	assertScanned(t, "- 1 2", "-", SYMBOL, 1, 1, 1, 1)
	// - sad case: a minus mid-number
	assertScanFailed(t, "1-2", "Error:1,2: invalid number format (minus can only appear at the beginning of a number)")
	assertScanFailed(t, "-", "Error:1,1: EOF")
}
