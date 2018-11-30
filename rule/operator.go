package rule

import (
	"encoding/json"
)

// The operator is the representation of an operation to be performed
// on some given set of operands.  Some Exprs are operators, but Value
// and Param types are not.
//
// Operators have a "kind" (the name of the operation), and a slice of
// operands (which are themselves Exprs).  Operands are added to an
// operator by means of the PushExpr call.
type operator struct {
	kind     string
	contract Contract
	operands []Expr
}

// PushExpr is used to add an Expr as an operand to the operator.
// Each call to PushExpr will check the new operand against the
// operators Contract, in the event that the new operand does not
// fulfil the appropriate Term of the Contract, and error will be
// returned. This causes operator to implement the Expr interface, and
// all operation expressions use it indirectly to do the same.
func (o *operator) PushExpr(e Expr) error {
	pos := len(o.operands)
	term, err := o.contract.GetTerm(pos, o.kind)
	if err != nil {
		return err
	}
	te := e.(TypedExpression)
	if !term.IsFulfilledBy(te) {
		return TypeError{
			OpCode:       o.kind,
			ErrorPos:     pos + 1,
			ExpectedType: term.Type,
			ReceivedType: te.Contract().ReturnType,
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
	}
	if pos < minPos {
		return ArityError{MinPos: minPos, ErrorPos: pos, OpCode: o.kind}
	}
	return nil
}

// Contract returns the operators Contract.  This makes operator
// implement the TypedExpression interface.  Its intent is to allow
// all operator Expr types to implement that interface (indirectly).
func (o *operator) Contract() Contract {
	return o.contract
}

func (o *operator) Same(c ComparableExpression) bool {
	if o.kind == c.GetKind() {
		o2, ok := c.(operander)
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

//
func (o *operator) GetKind() string {
	return o.kind
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

	tmpExpr, err := GetOperatorExpr(node.Kind)
	if err != nil {
		return err
	}
	o.kind = node.Kind
	o.contract = tmpExpr.(TypedExpression).Contract()

	for _, expr := range node.Operands.Exprs {
		o.PushExpr(expr)
	}

	return nil
}

func (o *operator) MarshalJSON() ([]byte, error) {
	var op struct {
		Kind     string `json:"kind"`
		Operands []Expr `json:"operands"`
	}

	op.Kind = o.kind
	op.Operands = o.operands
	return json.Marshal(&op)
}

func (o *operator) Eval(params Params) (*Value, error) {
	return nil, nil
}

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
