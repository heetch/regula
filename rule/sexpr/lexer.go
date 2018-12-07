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

// String returns a human readable description of the token type.
// This makes Token comply with the Stringer interface.
func (t Token) String() string {
	switch t {
	case EOF:
		return "end of file"
	case WHITESPACE:
		return "white space"
	case LPAREN: // (
		return "left parenthesis"
	case RPAREN: // )
		return "right parenthesis"
	case STRING:
		return "string"
	case NUMBER:
		return "number"
	case BOOL:
		return "Boolean"
	case COMMENT:
		return "comment"
	case SYMBOL:
		return "symbol"
	}
	return "invalid symbol"
}

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

// The lexicalElement is the structure returned by the Scanner and unit iterated over by the Parser.
type lexicalElement struct {
	Token   Token  // The Token indicates the type of lexical element represented
	Literal string // The Literal stores the string that was read
	// from the source code that is associated with
	// the lexicalElement.
	StartByte       int
	EndByte         int
	StartChar       int
	EndChar         int
	StartLine       int
	EndLine         int
	StartCharInLine int
	EndCharInLine   int
}

func newLE(tok Token, lit string) *lexicalElement {
	return &lexicalElement{Token: tok, Literal: lit}
}

// markStart records the start position of a lexical element
func (le *lexicalElement) markStart(startByte, startChar, startLine, startCharInLine int) {
	le.StartByte = startByte
	le.StartChar = startChar
	le.StartLine = startLine
	le.StartCharInLine = startCharInLine
}

// markEnd records the end position of a lexical element
func (le *lexicalElement) markEnd(endByte, endChar, endLine, endCharInLine int) {
	le.EndByte = endByte
	le.EndChar = endChar
	le.EndLine = endLine
	le.EndCharInLine = endCharInLine
}

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
func (s *Scanner) Scan() (*lexicalElement, error) {
	le := newLE(EOF, "")
	s.markStart(le)
	rn, err := s.readRune()
	if err != nil {
		s.markEnd(le)
		se := err.(*ScanError)
		if se.EOF {

			return le, nil
		}
		return le, err
	}
	switch {
	case isLParen(rn):
		le.Token = LPAREN
		le.Literal = "("
		s.markEnd(le)
		return le, nil
	case isRParen(rn):
		le.Token = RPAREN
		le.Literal = ")"
		s.markEnd(le)
		return le, nil
	case isWhitespace(rn):
		s.unreadRune(rn)
		return s.scanWhitespace()
	case isString(rn):
		return s.scanString()
	case isNumber(rn):
		s.unreadRune(rn)
		return s.scanNumber()
	case isBool(rn):
		return s.scanBool()
	case isComment(rn):
		return s.scanComment()
	case isSymbol(rn):
		s.unreadRune(rn)
		return s.scanSymbol()
	}

	le.Literal = string(rn)
	s.markEnd(le)
	return le, s.newScanError("Illegal character scanned")
}

func (s *Scanner) markStart(le *lexicalElement) {
	le.markStart(s.byteCount, s.charCount, s.lineCount, s.lineCharCount)
}

func (s *Scanner) markEnd(le *lexicalElement) {
	le.markEnd(s.byteCount, s.charCount, s.lineCount, s.lineCharCount)
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
func (s *Scanner) scanWhitespace() (*lexicalElement, error) {
	le := newLE(WHITESPACE, "")
	s.markStart(le)
	var b bytes.Buffer
	for {
		rn, err := s.readRune()
		if err != nil {
			le.Literal = b.String()
			s.markEnd(le)
			se := err.(*ScanError)
			if se.EOF {
				// We'll get EOF next time we try to
				// read a rune anyway, so we don't
				// have to care about it here, which
				// simplifies things.
				return le, nil

			}
			return le, err

		}
		if !isWhitespace(rn) {
			s.unreadRune(rn)
			break
		}
		b.WriteRune(rn)
	}
	le.Literal = b.String()
	s.markEnd(le)
	return le, nil
}

// scanString returns the contents of single, contiguous, double-quote delimited string constant.
func (s *Scanner) scanString() (*lexicalElement, error) {
	var b bytes.Buffer
	le := newLE(STRING, "")
	s.markStart(le)
	escape := false
	for {
		rn, err := s.readRune()
		if err != nil {
			s.markEnd(le)
			le.Literal = b.String()
			se := err.(*ScanError)
			if se.EOF {
				// we reached the end of the file
				// without seeing a terminator, that's
				// an error.
				return le, s.newScanError("unterminated string constant")
			}
			return le, err
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
	le.Literal = b.String()
	s.markEnd(le)
	return le, nil
}

// scanNumber scans a contiguous string representing a number.  As we
// have to handle the negative numeric form, it's possible that the
// '-' rune can prefix a number.  This is problematic because '-' can
// also be a symbol referring to the arithmetic operation "minus" -
// that confusion is resolved by scanNumber, and should it consider
// the latter case to be true it will return a SYMBOL rather than a
// NUMBER.
func (s *Scanner) scanNumber() (*lexicalElement, error) {
	var b bytes.Buffer
	le := newLE(NUMBER, "")
	s.markStart(le)

	// We can be certain this isn't EOF because we will already
	// have read and unread the rune before arriving here.
	rn, err := s.readRune()
	if err != nil {
		// Something drastic happened, because we read this fine the first time.
		s.markEnd(le)
		return le, err
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

		if err != nil {
			le.Literal = b.String()
			s.markEnd(le)
			se := err.(*ScanError)
			if se.EOF {
				// In reality having '-' as the final
				// symbol in a stream is never useful,
				// but this is the sort of error we
				// should catch in the Parser, not the
				// scanner.
				le.Token = SYMBOL
				return le, nil
			}
			return le, err
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
			le.Literal = b.String()
			le.Token = SYMBOL
			s.markEnd(le)
			return le, nil
		}
	}

	// OK, let's scan the rest of the number...

	for {
		rn, err := s.readRune()
		if err != nil {
			le.Literal = b.String()
			s.markEnd(le)
			se := err.(*ScanError)
			if se.EOF {
				// EOF is a valid terminator for a number
				return le, nil
			}
			return le, err
		}
		if rn == '-' {
			// As we said before '-' can't appear in the
			// body of a number, this is an error.
			le.Literal = b.String()
			s.markEnd(le)
			return le, s.newScanError("invalid number format (minus can only appear at the beginning of a number)")
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
	le.Literal = b.String()
	s.markEnd(le)
	return le, nil
}

// scanBool scans the contiguous characters following the '#' symbol,
// it they are either 'true', or 'false' a BOOL is returned, otherwise
// an ScanError will be returned.
func (s *Scanner) scanBool() (*lexicalElement, error) {
	var b bytes.Buffer
	le := newLE(BOOL, "")
	s.markStart(le)
	for {
		rn, err := s.readRune()
		if err != nil {
			se := err.(*ScanError)
			if se.EOF {
				// EOF is a valid terminator for a boolean
				break
			}
			le.Literal = b.String()
			s.markEnd(le)
			return le, err
		}

		// isSymbol is handy shorthand for "it's not anything else"
		if !isSymbol(rn) {
			s.unreadRune(rn)
			break
		}
		b.WriteRune(rn)
	}

	symbol := b.String()
	le.Literal = symbol
	s.markEnd(le)

	if symbol == "true" || symbol == "false" {
		return le, nil
	}
	if len(symbol) > 0 {
		return le, s.newScanError(fmt.Sprintf("invalid boolean: %s", symbol))
	}
	return le, s.newScanError("invalid boolean")
}

// scanComment will scan to the end of the current line, consuming any and all chars prior to '\n'.
func (s *Scanner) scanComment() (*lexicalElement, error) {
	var b bytes.Buffer
	le := newLE(COMMENT, "")
	s.markStart(le)

	for {
		rn, err := s.readRune()
		if err != nil {
			se := err.(*ScanError)
			if se.EOF {
				// EOF is a valid terminator for a Comment
				break
			}
			le.Literal = b.String()
			s.markEnd(le)
			return le, err
		}

		if rn == '\n' {
			break
		}

		b.WriteRune(rn)
	}
	le.Literal = b.String()
	s.markEnd(le)
	return le, nil
}

// scanSymbol scans a contiguous block of symbol characters. Any non-symbol character will terminate it.
func (s *Scanner) scanSymbol() (*lexicalElement, error) {
	var b bytes.Buffer

	le := newLE(SYMBOL, "")
	s.markStart(le)

	for {
		rn, err := s.readRune()
		if err != nil {
			se := err.(*ScanError)
			if se.EOF {
				// EOF is a valid terminator for a Symbol
				break
			}
			le.Literal = b.String()
			s.markEnd(le)
			return le, err
		}
		// Again, we have to special case '-', which can't start a symbol, but can appear in it.
		// Likewise numbers.
		if !(isSymbol(rn) || rn == '-' || unicode.IsDigit(rn)) {
			s.unreadRune(rn)
			break
		}
		b.WriteRune(rn)
	}
	le.Literal = b.String()
	s.markEnd(le)
	return le, nil
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
