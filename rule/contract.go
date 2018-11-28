package rule

import "fmt"

// Type represents the type of a typed expression.  Any expression has
// a return type, and some expressions also receive typed parameters.
type Type int

// String returns a string representation of the Type.  This makes
// Type implement the Stringer interface.
func (t Type) String() string {
	switch t {
	case BOOLEAN:
		return "BOOLEAN"
	case STRING:
		return "STRING"
	case INTEGER:
		return "INTEGER"
	case FLOAT:
		return "FLOAT"
	case NUMBER:
		return "NUMBER"
	case ANY:
		return "ANY"
	}
	return "Invalid Type"
}

// These constants represent the complete set of abstract types usable
// in expressions within Regula.  Not that these abstract types don't
// necessary equate directly to concrete types that you might think
// about within Go.  In particular "NUMBER" and "ANY" exist to define
// parameters that do not have to be of a fixed type.
const (
	BOOLEAN Type = iota
	STRING
	INTEGER
	FLOAT
	NUMBER // A special type that can be promoted to INTEGER or FLOAT
	ANY    // A special type that can be promoted to any other.
)

// String returns a human readable representation of the Type.  This
// makes Type implement the Stringer interface.
func (t Type) String() string {
	switch t {
	case BOOLEAN:
		return "Boolean"
	case STRING:
		return "String"
	case INTEGER:
		return "Integer"
	case FLOAT:
		return "Float"
	case NUMBER:
		return "Number"
	case ANY:
		return "Any"
	}
	return "invalid type"
}

// Cardinality expresses the number of times a term might be repeated
// in an expression.
type Cardinality int

// These constants define the complete set of cardinality possible for
// terms of a contract.  We can expect a term to occur exactly once
// (ONE), or zero, one or many times (MANY).
const (
	MANY = 0
	ONE  = 1
)

// Term represents the contract for an operand, or for a repeated
// sequence of operands.  A term can be fulfilled by a Value, Param or
// Expr, so long as their own Contract specifies a ReturnType that
// matches the Term's Type, or can be promoted to that Type.  Every
// Term also has Cardinality.
type Term struct {
	Type        Type
	Cardinality Cardinality
	Min         int // For Terms with Cardinality == MANY, we can specify a minimum number
}

// IsFulfilledBy returns true when a provided Expr has a return type
// that matches the Term's Type.
func (t Term) IsFulfilledBy(e Expr) bool {
	rt := e.Contract().ReturnType
	// Switch to handle promotable abstract types.
	switch t.Type {
	case ANY:
		return true
	case NUMBER:
		return rt == INTEGER || rt == FLOAT
	}
	return rt == t.Type
}

//Equal returns true when two Terms are identical
func (t Term) Equal(other Term) bool {
	return t.Type == other.Type && t.Cardinality == other.Cardinality
}

// A Contract declares the Type compatibility of an expression.  Every
// expression has a ReturnType (it's value type, in place when
// evaluated) and zero, one or many Terms.  Each Term is in turn
// typed, and has a defined cardinality.
type Contract struct {
	OpCode     string
	ReturnType Type
	Terms      []Term
}

// GetTerm returns the Term of a contract that matches a particular position.  If no Term is available for a position then an ArityError will be returned
func (c *Contract) GetTerm(pos int) (Term, error) {
	extent := len(c.Terms)
	if pos < extent {
		return c.Terms[pos], nil
	}
	lastTerm := c.Terms[extent-1]
	if lastTerm.Cardinality == MANY {
		return lastTerm, nil
	}
	return lastTerm, ArityError{OpCode: c.OpCode, ErrorPos: pos + 1, MaxPos: extent}
}

//Equal returns true when two contracts are identical.
func (c Contract) Equal(other Contract) bool {
	if c.ReturnType != other.ReturnType {
		return false
	}
	if len(c.Terms) != len(other.Terms) {
		return false
	}
	for i, ct := range c.Terms {
		if !ct.Equal(other.Terms[i]) {
			return false
		}
	}
	return true
}

// A TypedExpression is an expression that declares the Type Contract
// it makes with the context in which it appears, and with any
// sub-expressions that it contains.  This Contract can be inspected
// by calling the Contract method of the TypedExpression interface.
type TypedExpression interface {
	Contract() Contract
}

// GetOperatorExpr returns an Expr that matches the provided operator
// name. If no matching expression exists, and error will be returned.
func GetOperatorExpr(name string) (Expr, error) {
	switch name {
	case "eq":
		return newExprEq(), nil
	case "not":
		return newExprNot(), nil
	case "and":
		return newExprAnd(), nil
	case "or":
		return newExprOr(), nil
	case "in":
		return newExprIn(), nil

	}
	return nil, fmt.Errorf("No operator Expression called %q exists", name)
}
