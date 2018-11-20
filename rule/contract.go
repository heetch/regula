package rule

// Type represents the type of a typed expression.  Any expression has
// a return type, and some expressions also receive typed parameters.
type Type int

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

type Term struct {
	Type        Type
	Cardinality Cardinality
}

// IsFulfilledBy returns true when a provided TypedExpression has a return type
// that matches the Term's Type.
func (t Term) IsFulfilledBy(te TypedExpression) bool {
	return te.Contract().ReturnType == t.Type
}

// A Contract declares the Type compatibility of an expression.  Every
// expression has a ReturnType (it's value type, in place when
// evaluated) and zero, one or many Terms.  Each Term is in turn
// typed, and has a defined cardinality.
type Contract struct {
	ReturnType Type
	Terms      []Term
}

// A TypedExpression is an expression that declares the Type Contract
// it makes with the context in which it appears, and with any
// sub-expressions that it contains.  This Contract can be inspected
// by called the Contract method of the TypedExpression interface.
type TypedExpression interface {
	Contract() Contract
}
