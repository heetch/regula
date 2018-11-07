package sexpr

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
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
	r                     *bufio.Reader
	byteCount             int
	charCount             int
	lineCount             int
	lineCharCount         int
	previousLineCharCount int
}

// NewScanner wraps a Scanner around the provided io.Reader so that we
// might scan lexical tokens for the rule symbolic expression language
// from it.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r:         bufio.NewReader(r),
		lineCount: 1,
	}
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
	case isWhitespace(rn):
		err := s.unreadRune(rn)
		if err != nil {
			return EOF, string(rn), err
		}
		return s.scanWhitespace()
	case isString(rn):
		return s.scanString()
	}

	return EOF, string(rn), s.newScanError("Illegal character scanned")
}

func (s *Scanner) readRune() (rune, error) {
	rn, size, err := s.r.ReadRune()
	// EOF is a special case, it shouldn't affect counts
	if err == io.EOF {
		return rn, s.eof()
	}
	// We need to update the counts correctly before considering
	// any error, so that the data embedded in the ScanError is
	// correct.
	s.byteCount += size
	s.charCount++
	s.lineCharCount++
	if rn == '\n' {
		// DOS/Windows encoding does \n\r for new lines, but
		// we can ignore the \r and still get the right
		// result.
		s.lineCount++
		// Store the previous line char count in case we unread
		s.previousLineCharCount = s.lineCharCount
		// it's char zero, the next readRune should take us to 1
		s.lineCharCount = 0
	}
	if err != nil {
		return rn, s.newScanError(err.Error())
	}
	return rn, nil
}

func (s *Scanner) unreadRune(rn rune) error {
	err := s.r.UnreadRune()
	if err != nil {
		return s.newScanError(err.Error())
	}
	// Decrement counts after the unread is complete
	s.byteCount -= utf8.RuneLen(rn)
	s.charCount--
	s.lineCharCount--
	if rn == '\n' {
		s.lineCount--
		s.lineCharCount = s.previousLineCharCount
		s.previousLineCharCount--
	}

	return nil
}

// scanWhitespace scans a contiguous sequence of whitespace
// characters.  Note that this will consume newlines as it goes,
// lexically speaking they're insignificant to the language.
func (s *Scanner) scanWhitespace() (Token, string, error) {
	var b bytes.Buffer
	for {
		rn, err := s.readRune()
		if err != nil {
			se := err.(*ScanError)
			if se.EOF {
				// We'll get EOF next time we try to
				// read a rune anyway, so we don't
				// have to care about it here, which
				// simplifies things.
				return WHITESPACE, b.String(), nil

			}
			return WHITESPACE, b.String(), err

		}
		if !isWhitespace(rn) {
			err = s.unreadRune(rn)
			if err != nil {
				return WHITESPACE, b.String(), err
			}
			break
		}
		b.WriteRune(rn)
	}
	return WHITESPACE, b.String(), nil
}

// scanString returns the contents of single, contiguous, double-quote delimited string constant.
func (s *Scanner) scanString() (Token, string, error) {
	var b bytes.Buffer
	escape := false
	for {
		rn, err := s.readRune()
		if err != nil {
			se := err.(*ScanError)
			if se.EOF {
				// we reached the end of the file
				// without seeing a terminator, that's
				// an error.
				return STRING, b.String(), s.newScanError("unterminated string constant")
			}
			return STRING, b.String(), err
		}
		if escape {
			b.WriteRune(rn)
			escape = false
			continue
		}
		if isString(rn) {
			break
		}
		escape = rn == '\\'
		if !escape {
			b.WriteRune(rn)
		}
	}
	return STRING, b.String(), nil
}

//
func (s *Scanner) newScanError(message string) *ScanError {
	return &ScanError{
		Byte:       s.byteCount,
		Char:       s.charCount,
		Line:       s.lineCount,
		CharInLine: s.lineCharCount,
		msg:        message,
	}
}

func (s *Scanner) eof() *ScanError {
	err := s.newScanError("EOF")
	err.EOF = true
	return err
}

type ScanError struct {
	Byte       int
	Char       int
	Line       int
	CharInLine int
	msg        string
	EOF        bool
}

//
func (se ScanError) Error() string {
	return fmt.Sprintf("Error:%d,%d: %s", se.Line, se.CharInLine, se.msg)
}
