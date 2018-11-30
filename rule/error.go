package rule

import "fmt"

// ArityError captures structured information about an attempt to push
// an expression to an operator when no further expression is
// expected.
type ArityError struct {
	OpCode   string
	MaxPos   int
	ErrorPos int
}

//Error returns a string representation of the ArityError.  This makes
//ArityError implement the Error interface.
func (ae ArityError) Error() string {
	return fmt.Sprintf("attempted to pass an argument in position %d to %q operator, which only accepts %d arguments", ae.ErrorPos, ae.OpCode, ae.MaxPos)
}
