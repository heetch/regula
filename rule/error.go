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

// Error returns a string representation of the ArityError.  This makes
// ArityError implement the Error interface.
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

// Error returns a string representation of the TypeError.  This makes
// TypeError implement the Error interface.
func (te TypeError) Error() string {
	return fmt.Sprintf(`attempt to call %q with a %s in position %d, but it requires a %s`,
		te.OpCode, te.ReceivedType, te.ErrorPos, te.ExpectedType)
}

// HomogeneousTypeError encapsulates information relevant to an
// attempt to provide multiple Exprs with incompatible types as
// operands to an operator, which all match the same Term of the
// Contract, where that Term has the Cardinality "MANY".
type HomogeneousTypeError struct {
	OpCode       string
	ErrorPos     int
	HomoStartPos int
	ExpectedType Type
	ReceivedType Type
}

// Error returns a string representation of the HomogeneousTypeError.   This makes HomogeneousTypeError implement the Error interface.
func (hte HomogeneousTypeError) Error() string {
	return fmt.Sprintf(`attempt to call %q with a %s in position %d, but all arguments after position %d must be of the same type, and you previously passed %s.`, hte.OpCode, hte.ReceivedType, hte.ErrorPos, hte.HomoStartPos, hte.ExpectedType)
}

type HomogeneousBodyTypeError struct {
	HomogeneousTypeError
}

// Error returns a string representation of the HomogeneousBodyTypeError.   This makes HomogeneousBodyTypeError implement the Error interface.
func (hbte HomogeneousBodyTypeError) Error() string {
	return fmt.Sprintf(`attempt to call %q with a %s in position %d, but all body arguments must be of the same type, and you previously passed %s.`, hbte.OpCode, hbte.ReceivedType, hbte.ErrorPos, hbte.ExpectedType)
}
