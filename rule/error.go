package rule

import "fmt"

// ArityError captures structured information about an attempt to push
// an expression to an operator when no further expression is
// expected.
type ArityError struct {
	OpCode   string
	MinPos   int
	MaxPos   int
	ErrorPos int
}

//Error returns a string representation of the ArityError.  This makes
//ArityError implement the Error interface.
func (ae ArityError) Error() string {
	const template = "attempted to call %q with %d %s, but it requires %d %s"
	argNoun := func(count int) string {
		if count == 1 {
			return "argument"
		}
		return "arguments"
	}

	var pos int
	if ae.MaxPos > 0 && ae.ErrorPos > ae.MaxPos {
		pos = ae.MaxPos
	} else {
		pos = ae.MinPos
	}
	return fmt.Sprintf(template, ae.OpCode, ae.ErrorPos, argNoun(ae.ErrorPos), pos, argNoun(pos))
}

// TypeError encapsulates information about an attempt to call
// PushExpr on an operator with an Expr whose contracts ReturnType
// doesn't fulfil the positional Term of the operator into which
// it being pushed.
type TypeError struct {
	OpCode       string
	ErrorPos     int
	ExpectedType Type
	ReceivedType Type
}

// Error returns a string representation of the TypeError.  This make
// TypeError implement the Error interface.
func (te TypeError) Error() string {
	return fmt.Sprintf(`attempt to call %q with a %s in position %d, but it requires a %s`,
		te.OpCode, te.ReceivedType, te.ErrorPos, te.ExpectedType)
}
