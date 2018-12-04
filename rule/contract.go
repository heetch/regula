package rule

// Type represents the type of a typed expression.  Any expression has
// a return type, and some expressions also receive typed parameters.
type Type int

//String returns a human readable representation of the Type.  This
//makes Type implement the Stringer interface.
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

// A Contract declares the Type compatibility of an expression.  Every
// expression has a ReturnType (it's value type, in place when
// evaluated) and zero, one or many Terms.  Each Term is in turn
// typed, and has a defined cardinality.
type Contract struct {
	ReturnType Type
	Terms      []Term
}

//
func (c *Contract) GetTerm(pos int, opCode string) (Term, error) {
	extent := len(c.Terms)
	if pos < extent {
		return c.Terms[pos], nil
	}
	lastTerm := c.Terms[extent-1]
	if lastTerm.Cardinality == MANY {
		return lastTerm, nil
	}
	return lastTerm, ArityError{OpCode: opCode, ErrorPos: pos + 1, MaxPos: extent}

}
