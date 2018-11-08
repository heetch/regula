package sexpr

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// Tokens are the fundamental identifier of the lexical scanner.
// Every scanned element will be assigned a token type.
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
	return r == '-' || unicode.IsDigit(r)
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
		s.unreadRune(rn)
		return s.scanWhitespace()
	case isString(rn):
		return s.scanString()
	case isNumber(rn):
		s.unreadRune(rn)
		return s.scanNumber()
	}

	return EOF, string(rn), s.newScanError("Illegal character scanned")
}

// readRune pulls the next rune from the input sequence.
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

// unreadRune puts the last readRune back on the buffer and resets the
// counters.  It requires that the rune to be unread is passed, as we
// need to know the byte size of the rune.
func (s *Scanner) unreadRune(rn rune) {
	err := s.r.UnreadRune()
	if err != nil {
		// This means something truly awful happened!
		panic(err.Error())
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
			s.unreadRune(rn)
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

// scanNumber scans a contiguous string representing a number.  As we
// have to handle the negative numeric form, it's possible that the
// '-' rune can prefix a number.  This is problematic because '-' can
// also be a symbol referring to the arithmetic operation "minus" -
// that confusion is resolved by scanNumber, and should it consider
// the latter case to be true it will return a SYMBOL rather than a
// NUMBER.
func (s *Scanner) scanNumber() (Token, string, error) {
	var b bytes.Buffer

	// We can be certain this isn't EOF because we will already
	// have read and unread the rune before arriving here.
	rn, err := s.readRune()
	if err != nil {
		// Something drastic happened, because we read this fine the first time.
		return NUMBER, "", err
	}

	// Whatever happens we'll want the rune.
	b.WriteRune(rn)

	// Deal with the first rune.  Numbers have special rules about
	// the first rune, specifically it, and only it, may be the
	// minus symbol. When we loop later any occurrence of '-' will
	// be an error.
	if rn == '-' {

		// Now we look ahead to see if a number is coming, if
		// its anything else then this isn't a negative
		// number, but some other form.
		rn, err := s.readRune()

		// EOF would leave us with a '-' on its own.
		// This is never valid, so we can just promote
		// the error without bothering to check.
		if err != nil {
			return NUMBER, b.String(), err
		}

		// We've stored the rune, and we know we'll want to
		// unread whatever happens, so lets just do that now.
		s.unreadRune(rn)

		// If the next rune isn't a digit then we're going to
		// assume this is the minus operator and return '-' as
		// a symbol instead of a number.  There are still
		// cases where this wouldn't be valid, but they're all
		// errors and we'll leave that for the Parser to
		// handle.
		if !unicode.IsDigit(rn) {
			return SYMBOL, b.String(), nil
		}
	}

	// OK, let's scan the rest of the number...

	for {
		rn, err := s.readRune()
		if err != nil {
			se := err.(*ScanError)
			if se.EOF {
				// EOF is a valid terminator for a number
				return NUMBER, b.String(), nil
			}
			return NUMBER, b.String(), err
		}
		if rn == '-' {
			// As we said before '-' can't appear in the
			// body of a number, this is an error.
			return NUMBER, b.String(), s.newScanError("invalid number format (minus can only appear at the beginning of a number)")
		}

		// Valid number parts are written to the buffer
		if isNumber(rn) || rn == '.' {
			b.WriteRune(rn)
			continue
		}
		// we hit a terminating character, end the number here.
		s.unreadRune(rn)
		break
	}
	return NUMBER, b.String(), nil
}

// newScanError returns a ScanError initialised with the current
// positional information of the Scanner.
func (s *Scanner) newScanError(message string) *ScanError {
	return &ScanError{
		Byte:       s.byteCount,
		Char:       s.charCount,
		Line:       s.lineCount,
		CharInLine: s.lineCharCount,
		msg:        message,
	}
}

// eof returns a ScanError, initialised with the current positional
// information of the Scanner, and with it's EOF field set to True.
func (s *Scanner) eof() *ScanError {
	err := s.newScanError("EOF")
	err.EOF = true
	return err
}

// ScanError is a type that implements the Error interface, but adds
// additional context information to errors that can be inspected.  It
// is intended to be used for all errors emerging from the Scanner.
type ScanError struct {
	Byte       int
	Char       int
	Line       int
	CharInLine int
	msg        string
	EOF        bool
}

// Error makes ScanError comply with the Error interface.  It returns
// a string representation of the ScanError including it's message and
// some human readable position information.
func (se ScanError) Error() string {
	return fmt.Sprintf("Error:%d,%d: %s", se.Line, se.CharInLine, se.msg)
}
