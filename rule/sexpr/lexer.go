package sexpr

import (
	"bufio"
	"io"
	"unicode"
)

type Token int

const (
	EOF Token = iota
	WHITESPACE
	LPAREN // (
	RPAREN // )
	STRING
	NUMBER
	BOOL
	COMMENT
	SYMBOL
)

// isWhitespace returns true if the rune is the first rune of a
// Whitespace sequence.
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

// isLParen returns true if the rune is a left parenthesis
func isLParen(r rune) bool {
	return r == '('
}

// isRParen returns true if the rune is a right parenthesis
func isRParen(r rune) bool {
	return r == ')'
}

// isString returns true if the rune is a double quote indicating the
// beginning of string
func isString(r rune) bool {
	return r == '"'
}

// isNumber returns true if the rune is an Arabic number or the minus
// symbol, indicating the beginning of a number.
func isNumber(r rune) bool {
	// Note, although we allow a number to contain a decimal
	// point, it can't start with one so we don't include that in
	// the predicate.
	return r == '-' || (r >= '0' && r <= '9')
}

// isBool returns true if the rune is the # (hash or octothorpe)
// indicating the beginning of a boolean.
func isBool(r rune) bool {
	return r == '#'
}

// isComment returns true if the rune is ; (semicolon) indicating the
// beginning of a comment.
func isComment(r rune) bool {
	return r == ';'
}

// isSymbol returns true if no other special character type matches
// the rune.
func isSymbol(r rune) bool {
	return !(isWhitespace(r) || isLParen(r) || isRParen(r) || isString(r) || isNumber(r) || isBool(r) || isComment(r))
}

// Scanner is a lexical scanner for extracting the lexical tokens from
// a string of characters in our rule symbolic expression language.
type Scanner struct {
	r             *bufio.Reader
	byteCount     int
	charCount     int
	lineCount     int
	lineCharCount int
}

// NewScanner wraps a Scanner around the provided io.Reader so that we
// might scan lexical tokens for the rule symbolic expression language
// from it.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next lexical token found in the Scanner's io.Reader.
func (s *Scanner) Scan() (Token, string, error) {
	rn, err := s.readRune()
	if err != nil {
		return EOF, "", s.newScanError(err.Error())
	}
	switch {
	case isLParen(rn):
		return LPAREN, "(", nil
	case isRParen(rn):
		return RPAREN, ")", nil
	}
	return EOF, string(rn), s.newScanError("Illegal character scanned")
}

func (s *Scanner) readRune() (rune, error) {
	rn, size, err := s.r.ReadRune()
	s.byteCount += size
	s.charCount++
	s.lineCharCount++
	if rn == '\n' {
		// DOS/Windows encoding does \n\r for new lines, but
		// we can ignore the \r and still get the right
		// result.
		s.lineCount++
		// it's char zero, the next readRune should take us to 1
		s.lineCharCount = 0
	}
	return rn, err

}

//
func (s *Scanner) newScanError(message string) *ScanError {
	return &ScanError{
		Byte:       s.byteCount,
		Char:       s.charCount,
		Line:       s.lineCount + 1,
		CharInLine: s.lineCharCount,
		msg:        message,
	}
}

type ScanError struct {
	Byte       int
	Char       int
	Line       int
	CharInLine int
	msg        string
}

//
func (se *ScanError) Error() string {
	return se.msg
}
