package sexpr

import (
	"fmt"

	"github.com/heetch/regula/rule"
)

// ErrType is a broad classification of an error condition by type.
// It is hoped that these types will help a human to understand the
// nature of the error they have created.
type ErrorType uint

const (
	ARITY ErrorType = iota
	TYPE
	LEX
	PARSE
	OTHER
)

// String returns a human readable representation of the ErrorType.  This
// makes ErrorType implement the Stringer interface
func (e ErrorType) String() string {
	switch e {
	case ARITY:
		return "Arity error"
	case TYPE:
		return "Type error"
	case LEX:
		return "Lexical error"
	case PARSE:
		return "Parsing error"
	case OTHER:
		return "Error"
	}
	return "Undefined Error"
}

// ParserError is a wrapper for any and all errors occurring the
// symbolic expression parser that are to be made visible to the human
// submitting rules to the parser.  Its intent is to pass back enough
// information to the user interface that they can in term indicate to
// the human the exact location and nature of the issue in their code.
type ParserError struct {
	ErrorType       ErrorType
	Msg             string
	StartByte       int
	EndByte         int
	StartChar       int
	EndChar         int
	StartLine       int
	EndLine         int
	StartCharInLine int
	EndCharInLine   int
}

// newParseError take a lexicalElement and an error and creates a new
// ParseError that presents the error condition in the context of the
// lexical element.
func newParserError(le *lexicalElement, err error) ParserError {
	var errType ErrorType
	switch v := err.(type) {
	case ScanError:
		return ParserError{
			Msg:             v.msg,
			ErrorType:       LEX,
			StartByte:       v.Byte,
			EndByte:         v.Byte,
			StartChar:       v.Char,
			EndChar:         v.Char,
			StartLine:       v.Line,
			EndLine:         v.Line,
			StartCharInLine: v.CharInLine,
			EndCharInLine:   v.CharInLine,
		}
	case rule.ArityError:
		errType = ARITY

	case rule.TypeError:
		errType = TYPE
	default:
		errType = OTHER
	}

	return ParserError{
		Msg:             err.Error(),
		ErrorType:       errType,
		StartByte:       le.StartByte,
		EndByte:         le.EndByte,
		StartChar:       le.StartChar,
		EndChar:         le.EndChar,
		StartLine:       le.StartLine,
		EndLine:         le.EndLine,
		StartCharInLine: le.StartCharInLine,
		EndCharInLine:   le.EndCharInLine,
	}

}

// Error returns a human readable summary of the ParseError.  This
// make ParseError implement the Error interface.
func (p ParserError) Error() string {
	return fmt.Sprintf(
		"%d:%d: %s. %s.",
		p.StartLine, p.StartCharInLine, p.ErrorType, p.Msg)
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
