package rule

import (
	"encoding/json"
	"fmt"
)

// An Operator is an Expr that is also an Operander.
type Operator interface {
	Expr
	Operander
}

type OperatorConstructor func() Operator

var Operators = make(map[string]OperatorConstructor)

// GetOperator returns an Operator that matches the provided operator
// name. If no matching expression exists, an error will be returned.
func GetOperator(name string) (Operator, error) {
	opCon, ok := Operators[name]
	if !ok {
		return nil, fmt.Errorf("no operator Expression called %q exists", name)
	}
	return opCon(), nil
}

// The operator is the representation of an operation to be performed
// on some given set of operands.  Some Exprs are operators, but Value
// and Param types are not.
//
// Operators have a Contract , and a slice of operands (which are
// themselves Exprs).  Operands are added to an operator by means of
// the PushExpr call.
type operator struct {
	contract Contract
	operands []Expr
}

// consumeOperands allows us to populate the operator with Operands in
// line with its Contract.  Should one of the Exprs provided not match
// the terms of the Contract, this function will panic.
func (o *operator) consumeOperands(e ...Expr) {
	for _, ex := range e {
		o.pushExprOrPanic(ex)
	}
	o.finaliseOrPanic()
}

// PushExpr is used to add an Expr as an operand to the operator.
// Each call to PushExpr will check the new operand against the
// operators Contract, in the event that the new operand does not
// fulfil the appropriate Term of the Contract, an error will be
// returned. This causes operator to implement the Expr interface, and
// all operation expressions use it indirectly to do the same.
func (o *operator) PushExpr(e Expr) error {
	pos := len(o.operands)
	term, err := o.contract.GetTerm(pos)
	if err != nil {
		return err
	}
	if !term.IsFulfilledBy(e) {
		return TypeError{
			OpCode:       o.GetKind(),
			ErrorPos:     pos + 1,
			ExpectedType: term.Type,
			ReceivedType: e.Contract().ReturnType,
		}
	}
	o.operands = append(o.operands, e)
	return nil
}

// Finalise tells the operator that we're done pushing Exprs to use as
// operands, and in doing so gives the operator a chance to check that
// we've pushed enough operands to fulfil the operators Contract.  In
// the case that we haven't pushed enough operands, an ArityError will
// be returned.  This causes operator to implement the Expr interface,
// and indirectly all of the operation expressions use it to do the
// same.
func (o *operator) Finalise() error {
	pos := len(o.operands)
	extent := len(o.contract.Terms)
	lastTerm := o.contract.Terms[extent-1]
	minPos := extent
	if lastTerm.Cardinality == MANY {
		// First we subtract one, because MANY can mean zero,
		// then we add on the minimum required.
		minPos = (minPos - 1) + lastTerm.Min
		o.homogenise()
	}
	if pos < minPos {
		return ArityError{MinPos: minPos, ErrorPos: pos, OpCode: o.GetKind()}
	}

	return nil
}

// Contract returns the operators Contract.  This makes operator
// implement the TypedExpression interface.  Its intent is to allow
// all operator Expr types to implement that interface (indirectly).
func (o *operator) Contract() Contract {
	return o.contract
}

// Same returns true if the operator is the root of an abstract syntax
// tree that exactly matches the one described by a provided
// ComparableExpression.  This method is intended for use in test
// cases making assertions about other ways of generating an abstract
// syntax tree.  This makes operator implement the
// ComparableExpression interface.
func (o *operator) Same(c ComparableExpression) bool {
	if o.GetKind() == c.GetKind() {
		o2, ok := c.(Operander)
		if ok {
			ops := o2.Operands()
			len1 := len(o.operands)
			len2 := len(ops)
			if len1 != len2 {
				return false
			}
			for i := 0; i < len1; i++ {
				ce1 := o.operands[i].(ComparableExpression)
				ce2 := ops[i].(ComparableExpression)
				if !ce1.Same(ce2) {
					return false
				}
			}
			return true
		}
		return false
	}
	return false
}

// GetKind returns the kind of the operator.  This makes operator
// implement the ComparableExpression interface.
func (o *operator) GetKind() string {
	return o.contract.OpCode
}

func (o *operator) UnmarshalJSON(data []byte) error {
	var node struct {
		Kind     string
		Operands operands
	}

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	tmpExpr, err := GetOperator(node.Kind)
	if err != nil {
		return err
	}
	o.contract = tmpExpr.Contract()

	for _, expr := range node.Operands.Exprs {
		err := o.PushExpr(expr)
		if err != nil {
			return err
		}
	}
	err = o.Finalise()
	if err != nil {
		return err
	}

	return nil
}

func (o *operator) MarshalJSON() ([]byte, error) {
	var op struct {
		Kind     string `json:"kind"`
		Operands []Expr `json:"operands"`
	}

	op.Kind = o.GetKind()
	op.Operands = o.operands
	return json.Marshal(&op)
}

func (o *operator) Eval(params Params) (*Value, error) {
	return nil, nil
}

// Operands returns the operators Operands. This complies with the
// Operander interface.
func (o *operator) Operands() []Expr {
	return o.operands
}

// Support contract checking in the legacy Go interface for rule
// expression by panicking if something breaks the contract.  This
// works to the explicit assumption that developers won't release
// panicking code into production.
func (o *operator) pushExprOrPanic(e Expr) {
	if err := o.PushExpr(e); err != nil {
		panic(err.Error())
	}
}

// Support contract checking in the legacy Go interface for rule
// expression by panicking if something breaks the contract.  This
// works to the explicit assumption that developers won't release
// panicking code into production.
func (o *operator) finaliseOrPanic() {
	if err := o.Finalise(); err != nil {
		panic(err.Error())
	}
}

// homogenise will check all the Exprs passed for a Term with the
// Cardinality "MANY" and ensure that they are the same type,
// promoting the type where possible, and raising a
// HomogeneousTypeError where promotion is not possible.
func (o *operator) homogenise() error {
	// Only the last Term can be defined with Cardinality == MANY,
	// and therefore only that case can care about whether it has
	// homogeneous arguments or not.
	lastIndex := len(o.contract.Terms) - 1
	if lastIndex >= 0 {
		lastTerm := o.contract.Terms[lastIndex]
		// If Cardinality != MANY then we can't have mixed types to worry about.
		if lastTerm.Cardinality == MANY {
			// This starts at ANY, the most generic type,
			// so we can move towards something more
			// specific.
			var t Type = ANY
			var isFloat bool
			var isInt bool

			// Iterate over all operands that match the
			// Term with Cardinality == MANY
			for i := lastIndex; i < len(o.operands); i++ {
				opr := o.operands[i]
				rt := opr.Contract().ReturnType
				if rt != t {
					// If we don't yet have a concrete type
					if t == ANY {
						switch rt {
						case ANY:
							panic("Operand with ANY ReturnType - this should never happen")
						case NUMBER:
							panic("Operand with NUMBER ReturnType - this should never happen")
						case INTEGER:
							isInt = true
						case FLOAT:
							isFloat = true
						}
						// The concrete type should now be this type
						t = rt
						continue
					}

					// If we've already seen a concrete type
					switch rt {
					case ANY:
						// This should never happen!
						panic("Operand with ANY ReturnType - this should never happen")
					case NUMBER:
						// This should never happen!
						panic("Operand with NUMBER ReturnType - this should never happen")
					case INTEGER:
						isInt = true
						if t != FLOAT {
							return o.newHomogeneousTypeError(i+1, lastIndex+1, t, rt)
						}

					case FLOAT:
						isFloat = true
						if t != INTEGER {
							return o.newHomogeneousTypeError(i+1, lastIndex+1, t, rt)
						}
						t = FLOAT
					default:
						return o.newHomogeneousTypeError(i+1, lastIndex+1, t, rt)
					}
					// The concrete type should now be this type
				}

			}

			// Is this a numeric type?
			if isInt || isFloat {
				t = INTEGER
				// Float overrides Int when we have both.
				if isFloat {
					if isInt {
						// We saw some INTEGERs amongst the FLOATs so, we'll promote those all to FLOAT
						o.promoteToFloat(lastIndex)
					}
					t = FLOAT
				}
				// If this op has NUMBER as its
				// ReturnType in the Contract, we need
				// to convert that to something
				// concrete before it can be used as
				// an argument in another operator.
				if o.contract.ReturnType == NUMBER {
					o.contract.ReturnType = t
				}
			}
		}
	}
	return nil
}

// promoteToFloat wraps any operand, with an INTEGER ReturnType, that
// appears in a position managed by a Term with Cardinality=MANY, in a
// IntToFloat Expr.
func (o *operator) promoteToFloat(startIndex int) {
	for i := startIndex; i < len(o.operands); i++ {
		opr := o.operands[i]
		if opr.Contract().ReturnType == INTEGER {
			o.operands[i] = IntToFloat(opr)
		}
	}
}

//newHomogeneousTypeError returns a HomogeneousTypeError for the operator
func (o *operator) newHomogeneousTypeError(pos, startPos int, expected, received Type) HomogeneousTypeError {

	return HomogeneousTypeError{
		OpCode:       o.contract.OpCode,
		ErrorPos:     pos,
		HomoStartPos: startPos,
		ExpectedType: expected,
		ReceivedType: received,
	}

}
