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

func assertScannerScanned(t *testing.T, s *Scanner, expected lexicalElement) {
	le, err := s.Scan()
	require.NoError(t, err)
	require.Equalf(t, expected.Token, le.Token, "Token")
	require.Equalf(t, expected.Literal, le.Literal, "Literal")
	require.Equalf(t, expected.StartByte, le.StartByte, "StartByte")
	require.Equalf(t, expected.StartChar, le.StartChar, "StartChar")
	require.Equalf(t, expected.StartLine, le.StartLine, "StartLine")
	require.Equalf(t, expected.StartCharInLine, le.StartCharInLine, "StartCharInLine")
	require.Equalf(t, expected.EndByte, le.EndByte, "EndByte")
	require.Equalf(t, expected.EndChar, le.EndChar, "EndChar")
	require.Equalf(t, expected.EndLine, le.EndLine, "EndLine")
	require.Equalf(t, expected.EndCharInLine, le.EndCharInLine, "EndCharInLine")
}

func assertScanned(t *testing.T, input string, expected lexicalElement) {
	t.Run(fmt.Sprintf("Scan %s 0x%x", input, input), func(t *testing.T) {
		b := bytes.NewBufferString(input)
		s := NewScanner(b)
		assertScannerScanned(t, s, expected)
	})
}

func assertScannerScanFailed(t *testing.T, s *Scanner, message string) {
	_, err := s.Scan()
	require.EqualError(t, err, message)

}

func assertScanFailed(t *testing.T, input, message string) {
	t.Run(fmt.Sprintf("Scan should fail %s 0x%x", input, input), func(t *testing.T) {
		b := bytes.NewBufferString(input)
		s := NewScanner(b)
		assertScannerScanFailed(t, s, message)
	})

}

func TestScannerScanParenthesis(t *testing.T) {
	// Test L Parenthesis
	assertScanned(t, "(", lexicalElement{
		Literal:         "(",
		Token:           LPAREN,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// Test R Parenthesis
	assertScanned(t, ")", lexicalElement{
		Literal:         ")",
		Token:           RPAREN,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
}

func TestScannerScanWhiteSpace(t *testing.T) {
	// Test white-space
	assertScanned(t, " ", lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	assertScanned(t, "\t", lexicalElement{
		Literal:         "\t",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	assertScanned(t, "\r", lexicalElement{
		Literal:         "\r",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	assertScanned(t, "\n", lexicalElement{
		Literal:         "\n",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         1,
		EndChar:         1,
		EndLine:         2,
		EndCharInLine:   0,
	})
	assertScanned(t, "\v", lexicalElement{
		Literal:         "\v",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	assertScanned(t, "\f", lexicalElement{
		Literal:         "\f",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// Test contiguous white-space:
	// - terminated by EOF
	assertScanned(t, "  ", lexicalElement{
		Literal:         "  ",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         2,
		EndChar:         2,
		EndLine:         1,
		EndCharInLine:   2,
	})
	// - terminated by non white-space character.
	assertScanned(t, "  (", lexicalElement{
		Literal:         "  ",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         2,
		EndChar:         2,
		EndLine:         1,
		EndCharInLine:   2,
	})
}

func TestScannerScanString(t *testing.T) {
	// Test string:
	// - the empty string
	assertScanned(t, `""`, lexicalElement{
		Literal:         "",
		Token:           STRING,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         2,
		EndChar:         2,
		EndLine:         1,
		EndCharInLine:   2,
	})
	// - the happy case
	assertScanned(t, `"foo"`, lexicalElement{
		Literal:         "foo",
		Token:           STRING,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         5,
		EndChar:         5,
		EndLine:         1,
		EndCharInLine:   5,
	})
	// - an unterminated sad case
	assertScanFailed(t, `"foo`, "Error:1,4: unterminated string constant")
	// - happy case with escaped double quote
	assertScanned(t, `"foo\""`, lexicalElement{
		Literal:         `foo"`,
		Token:           STRING,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         7,
		EndChar:         7,
		EndLine:         1,
		EndCharInLine:   7,
	})
	// - sad case with escaped terminator
	assertScanFailed(t, `"foo\"`, "Error:1,6: unterminated string constant")
}

func TestScannerScanNumber(t *testing.T) {
	// Test number
	// - Single digit integer, EOF terminated
	assertScanned(t, "1", lexicalElement{
		Literal:         "1",
		Token:           NUMBER,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// - Single digit integer, terminated by non-numeric character
	assertScanned(t, "1)", lexicalElement{
		Literal:         "1",
		Token:           NUMBER,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// - Multi-digit integer, EOF terminated
	assertScanned(t, "998989", lexicalElement{
		Literal:         "998989",
		Token:           NUMBER,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         6,
		EndChar:         6,
		EndLine:         1,
		EndCharInLine:   6,
	})
	// - Negative multi-digit integer, EOF terminated
	assertScanned(t, "-100", lexicalElement{
		Literal:         "-100",
		Token:           NUMBER,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         4,
		EndChar:         4,
		EndLine:         1,
		EndCharInLine:   4,
	})
	// - Floating point number, EOF terminated
	assertScanned(t, "2.4", lexicalElement{
		Literal:         "2.4",
		Token:           NUMBER,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         3,
		EndChar:         3,
		EndLine:         1,
		EndCharInLine:   3,
	})
	// - long negative float, terminated by non-numeric character
	assertScanned(t, "-123.45456 ", lexicalElement{
		Literal:         "-123.45456",
		Token:           NUMBER,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: -0,
		EndByte:         10,
		EndChar:         10,
		EndLine:         1,
		EndCharInLine:   10,
	})
	// - special case: a "-" without a number following it (as per the minus operator)
	assertScanned(t, "- 1 2", lexicalElement{
		Literal:         "-",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// - sad case: a minus mid-number
	assertScanFailed(t, "1-2", "Error:1,2: invalid number format (minus can only appear at the beginning of a number)")
}

func TestScannerScanBool(t *testing.T) {
	// Happy cases
	// - true,  EOF Terminated
	assertScanned(t, "#true", lexicalElement{
		Literal:         "true",
		Token:           BOOL,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         5,
		EndChar:         5,
		EndLine:         1,
		EndCharInLine:   5,
	})
	// - false, newline terminated
	assertScanned(t, "#false\n", lexicalElement{
		Literal:         "false",
		Token:           BOOL,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         6,
		EndChar:         6,
		EndLine:         1,
		EndCharInLine:   7,
	})
	// Sad cases
	// - partial true
	assertScanFailed(t, "#tru ", "Error:1,4: invalid boolean: tru")
	// - partial false
	assertScanFailed(t, "#fa)", "Error:1,3: invalid boolean: fa")
	// - invalid
	assertScanFailed(t, "#1", "Error:1,1: invalid boolean")
	// - repeated signal character
	assertScanFailed(t, "##", "Error:1,1: invalid boolean")
	// - empty
	assertScanFailed(t, "#", "Error:1,1: invalid boolean")
}

func TestScannerScanComment(t *testing.T) {
	// Simple empty comment at EOF
	assertScanned(t, ";", lexicalElement{
		Literal:         "",
		Token:           COMMENT,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// Comment terminated by newline
	assertScanned(t, "; Foo\nbar", lexicalElement{
		Literal:         " Foo",
		Token:           COMMENT,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         6,
		EndChar:         6,
		EndLine:         2,
		EndCharInLine:   0,
	})
	// Comment containing Comment char
	assertScanned(t, ";Pants;On;Fire", lexicalElement{
		Literal:         "Pants;On;Fire",
		Token:           COMMENT,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         14,
		EndChar:         14,
		EndLine:         1,
		EndCharInLine:   14,
	})
	// Comment containing control characters
	assertScanned(t, `;()"-#1`, lexicalElement{
		Literal:         `()"-#1`,
		Token:           COMMENT,
		StartByte:       1,
		StartChar:       1,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         7,
		EndChar:         7,
		EndLine:         1,
		EndCharInLine:   7,
	})
}

func TestScannerScanSymbol(t *testing.T) {
	// Simple, single character identifier
	assertScanned(t, "a", lexicalElement{
		Literal:         "a",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// Fully formed symbol
	assertScanned(t, "abba-sucks-123_ok!", lexicalElement{
		Literal:         "abba-sucks-123_ok!",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         18,
		EndChar:         18,
		EndLine:         1,
		EndCharInLine:   18,
	})
	// Unicode in symbols
	assertScanned(t, "mötlěy_crü_sucks_more", lexicalElement{
		Literal:         "mötlěy_crü_sucks_more",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         24,
		EndChar:         21,
		EndLine:         1,
		EndCharInLine:   21,
	})
	// terminated by comment
	assertScanned(t, "bon;jovi is worse", lexicalElement{
		Literal:         "bon",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         3,
		EndChar:         3,
		EndLine:         1,
		EndCharInLine:   3,
	})
	// terminated by whitespace
	assertScanned(t, "van halen is the worst", lexicalElement{
		Literal:         "van",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         3,
		EndChar:         3,
		EndLine:         1,
		EndCharInLine:   3,
	})
	// terminated by control character
	assertScanned(t, "NoWayMichaelBolton)IsTheNadir", lexicalElement{
		Literal:         "NoWayMichaelBolton",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         18,
		EndChar:         18,
		EndLine:         1,
		EndCharInLine:   18,
	})
	// symbol starting with a non-alpha character
	assertScanned(t, "+", lexicalElement{
		Literal:         "+",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
	// actually handled by the number scan, but we'll check '-' all the same:
	assertScanned(t, "-", lexicalElement{
		Literal:         "-",
		Token:           SYMBOL,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 0,
		EndByte:         1,
		EndChar:         1,
		EndLine:         1,
		EndCharInLine:   1,
	})
}

// Scanner.Scan can scan a full symbollic expression sequence.
func TestScannerScanSequence(t *testing.T) {
	input := `
(and
  (= (+ 1 -1) 0)
  (= my-parameter "fudge sundae")) ; Crazy
`
	b := bytes.NewBufferString(input)
	s := NewScanner(b)
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "\n",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         1,
		EndChar:         1,
		EndLine:         2,
		EndCharInLine:   0,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "(",
		Token:           LPAREN,
		StartByte:       1,
		StartChar:       1,
		StartLine:       2,
		StartCharInLine: 0,
		EndByte:         2,
		EndChar:         2,
		EndLine:         2,
		EndCharInLine:   1,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "and",
		Token:           SYMBOL,
		StartByte:       2,
		StartChar:       2,
		StartLine:       2,
		StartCharInLine: 1,
		EndByte:         5,
		EndChar:         5,
		EndLine:         2,
		EndCharInLine:   5,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "\n  ",
		Token:           WHITESPACE,
		StartByte:       5,
		StartChar:       5,
		StartLine:       2,
		StartCharInLine: 6,
		EndByte:         8,
		EndChar:         8,
		EndLine:         3,
		EndCharInLine:   2,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "(",
		Token:           LPAREN,
		StartByte:       8,
		StartChar:       8,
		StartLine:       3,
		StartCharInLine: 2,
		EndByte:         9,
		EndChar:         9,
		EndLine:         3,
		EndCharInLine:   3,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "=",
		Token:           SYMBOL,
		StartByte:       9,
		StartChar:       9,
		StartLine:       3,
		StartCharInLine: 3,
		EndByte:         10,
		EndChar:         10,
		EndLine:         3,
		EndCharInLine:   4,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       10,
		StartChar:       10,
		StartLine:       3,
		StartCharInLine: 4,
		EndByte:         11,
		EndChar:         11,
		EndLine:         3,
		EndCharInLine:   5,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "(",
		Token:           LPAREN,
		StartByte:       11,
		StartChar:       11,
		StartLine:       3,
		StartCharInLine: 5,
		EndByte:         12,
		EndChar:         12,
		EndLine:         3,
		EndCharInLine:   6,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "+",
		Token:           SYMBOL,
		StartByte:       12,
		StartChar:       12,
		StartLine:       3,
		StartCharInLine: 6,
		EndByte:         13,
		EndChar:         13,
		EndLine:         3,
		EndCharInLine:   7,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       13,
		StartChar:       13,
		StartLine:       3,
		StartCharInLine: 7,
		EndByte:         14,
		EndChar:         14,
		EndLine:         3,
		EndCharInLine:   8,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "1",
		Token:           NUMBER,
		StartByte:       14,
		StartChar:       14,
		StartLine:       3,
		StartCharInLine: 8,
		EndByte:         15,
		EndChar:         15,
		EndLine:         3,
		EndCharInLine:   9,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       15,
		StartChar:       15,
		StartLine:       3,
		StartCharInLine: 9,
		EndByte:         16,
		EndChar:         16,
		EndLine:         3,
		EndCharInLine:   10,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "-1",
		Token:           NUMBER,
		StartByte:       16,
		StartChar:       16,
		StartLine:       3,
		StartCharInLine: 10,
		EndByte:         18,
		EndChar:         18,
		EndLine:         3,
		EndCharInLine:   12,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         ")",
		Token:           RPAREN,
		StartByte:       18,
		StartChar:       18,
		StartLine:       3,
		StartCharInLine: 12,
		EndByte:         19,
		EndChar:         19,
		EndLine:         3,
		EndCharInLine:   13,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       19,
		StartChar:       19,
		StartLine:       3,
		StartCharInLine: 13,
		EndByte:         20,
		EndChar:         20,
		EndLine:         3,
		EndCharInLine:   14,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "0",
		Token:           NUMBER,
		StartByte:       20,
		StartChar:       20,
		StartLine:       3,
		StartCharInLine: 14,
		EndByte:         21,
		EndChar:         21,
		EndLine:         3,
		EndCharInLine:   15,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         ")",
		Token:           RPAREN,
		StartByte:       21,
		StartChar:       21,
		StartLine:       3,
		StartCharInLine: 15,
		EndByte:         22,
		EndChar:         22,
		EndLine:         3,
		EndCharInLine:   16,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "\n  ",
		Token:           WHITESPACE,
		StartByte:       22,
		StartChar:       22,
		StartLine:       3,
		StartCharInLine: 17,
		EndByte:         25,
		EndChar:         25,
		EndLine:         4,
		EndCharInLine:   2,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "(",
		Token:           LPAREN,
		StartByte:       25,
		StartChar:       25,
		StartLine:       4,
		StartCharInLine: 2,
		EndByte:         26,
		EndChar:         26,
		EndLine:         4,
		EndCharInLine:   3,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "=",
		Token:           SYMBOL,
		StartByte:       26,
		StartChar:       26,
		StartLine:       4,
		StartCharInLine: 3,
		EndByte:         27,
		EndChar:         27,
		EndLine:         4,
		EndCharInLine:   4,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       27,
		StartChar:       27,
		StartLine:       4,
		StartCharInLine: 4,
		EndByte:         28,
		EndChar:         28,
		EndLine:         4,
		EndCharInLine:   5,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "my-parameter",
		Token:           SYMBOL,
		StartByte:       28,
		StartChar:       28,
		StartLine:       4,
		StartCharInLine: 5,
		EndByte:         40,
		EndChar:         40,
		EndLine:         4,
		EndCharInLine:   17,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       40,
		StartChar:       40,
		StartLine:       4,
		StartCharInLine: 17,
		EndByte:         41,
		EndChar:         41,
		EndLine:         4,
		EndCharInLine:   18,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "fudge sundae",
		Token:           STRING,
		StartByte:       42,
		StartChar:       42,
		StartLine:       4,
		StartCharInLine: 19,
		EndByte:         55,
		EndChar:         55,
		EndLine:         4,
		EndCharInLine:   32,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         ")",
		Token:           RPAREN,
		StartByte:       55,
		StartChar:       55,
		StartLine:       4,
		StartCharInLine: 32,
		EndByte:         56,
		EndChar:         56,
		EndLine:         4,
		EndCharInLine:   33,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         ")",
		Token:           RPAREN,
		StartByte:       56,
		StartChar:       56,
		StartLine:       4,
		StartCharInLine: 33,
		EndByte:         57,
		EndChar:         57,
		EndLine:         4,
		EndCharInLine:   34,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       57,
		StartChar:       57,
		StartLine:       4,
		StartCharInLine: 34,
		EndByte:         58,
		EndChar:         58,
		EndLine:         4,
		EndCharInLine:   35,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " Crazy",
		Token:           COMMENT,
		StartByte:       59,
		StartChar:       59,
		StartLine:       4,
		StartCharInLine: 36,
		EndByte:         66,
		EndChar:         66,
		EndLine:         5,
		EndCharInLine:   0,
	})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "",
		Token:           EOF,
		StartByte:       66,
		StartChar:       66,
		StartLine:       5,
		StartCharInLine: 0,
		EndByte:         66,
		EndChar:         66,
		EndLine:         5,
		EndCharInLine:   0,
	})
}

func TestScannerScanReturnsScanError(t *testing.T) {
	input := `
(= "toffee`
	b := bytes.NewBufferString(input)
	s := NewScanner(b)
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "\n",
		Token:           WHITESPACE,
		StartByte:       0,
		StartChar:       0,
		StartLine:       1,
		StartCharInLine: 1,
		EndByte:         1,
		EndChar:         1,
		EndLine:         2,
		EndCharInLine:   0})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "(",
		Token:           LPAREN,
		StartByte:       1,
		StartChar:       1,
		StartLine:       2,
		StartCharInLine: 0,
		EndByte:         2,
		EndChar:         2,
		EndLine:         2,
		EndCharInLine:   1})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         "=",
		Token:           SYMBOL,
		StartByte:       2,
		StartChar:       2,
		StartLine:       2,
		StartCharInLine: 1,
		EndByte:         3,
		EndChar:         3,
		EndLine:         2,
		EndCharInLine:   2})
	assertScannerScanned(t, s, lexicalElement{
		Literal:         " ",
		Token:           WHITESPACE,
		StartByte:       3,
		StartChar:       3,
		StartLine:       2,
		StartCharInLine: 2,
		EndByte:         4,
		EndChar:         4,
		EndLine:         2,
		EndCharInLine:   3})
	assertScannerScanFailed(t, s, "Error:2,10: unterminated string constant")
}
