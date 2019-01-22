package rule

import (
	"errors"
	"fmt"
)

func init() {
	Operators["eq"] = func() Operator { return newExprEq() }

}

/////////////////
// Eq operator //
/////////////////

type exprEq struct {
	operator
}

func newExprEq() *exprEq {
	return &exprEq{
		operator: operator{
			contract: Contract{
				OpCode:     "eq",
				ReturnType: BOOLEAN,
				Terms: []Term{
					{
						Type:        ANY,
						Cardinality: MANY,
						Min:         2,
					},
				},
			},
		},
	}
}

// Eq creates an expression that takes at least two operands and evaluates to true if all the operands are equal.
func Eq(vN ...Expr) Expr {
	e := newExprEq()
	e.consumeOperands(vN...)
	return e
}

func (n *exprEq) Eval(params Params) (*Value, error) {
	if len(n.operands) < 2 {
		return nil, errors.New("invalid number of operands in Eq func")
	}

	opA := n.operands[0]
	vA, err := opA.Eval(params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.operands); i++ {
		vB, err := n.operands[i].Eval(params)
		if err != nil {
			return nil, err
		}

		if !vA.Equal(vB) {
			return BoolValue(false), nil
		}
	}

	return BoolValue(true), nil
}

/////////////////
// LT operator //
/////////////////

type exprLT struct {
	operator
}

func newExprLT() *exprLT {
	return &exprLT{
		operator: operator{
			contract: Contract{
				OpCode:     "lt",
				ReturnType: BOOLEAN,
				Terms: []Term{
					{
						Type:        ANY,
						Cardinality: MANY,
						Min:         2,
					},
				},
			},
		},
	}
}

// LT creates an expression that takes at least two operands and
// evaluates to true if each successive operand has a lower value than
// the next.
func LT(vN ...Expr) Expr {
	e := newExprLT()
	e.consumeOperands(vN...)
	return e
}

// perform a less-than comparison on a sequence of integers
func (n *exprLT) integerLT(params Params) (*Value, error) {
	var i0, i1 int64
	var err error

	i0, err = exprToInt64(n.operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(n.operands); j++ {
		i1, err = exprToInt64(n.operands[j], params)
		if err != nil {
			return nil, err
		}
		if i0 >= i1 {
			return BoolValue(false), nil
		}
		i0 = i1
	}
	return BoolValue(true), nil
}

// perform a less-than comparison on a sequence of floats
func (n *exprLT) floatLT(params Params) (*Value, error) {
	var f0, f1 float64
	var err error

	f0, err = exprToFloat64(n.operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(n.operands); j++ {
		f1, err = exprToFloat64(n.operands[j], params)
		if err != nil {
			return nil, err
		}
		if f0 >= f1 {
			return BoolValue(false), nil
		}
		f0 = f1
	}
	return BoolValue(true), nil
}

// perform a less-than comparison on a sequence of strings
func (n *exprLT) stringLT(params Params) (*Value, error) {
	var s0, s1 string
	var err error

	s0, err = exprToString(n.operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(n.operands); j++ {
		s1, err = exprToString(n.operands[j], params)
		if err != nil {
			return nil, err
		}
		if s0 >= s1 {
			return BoolValue(false), nil
		}
		s0 = s1
	}
	return BoolValue(true), nil
}

func (n *exprLT) Eval(params Params) (*Value, error) {
	// Because of homogenisation during Parsing we know that all
	// operands have the same type.
	t := n.operands[0].Contract().ReturnType
	switch t {
	case INTEGER:
		return n.integerLT(params)
	case FLOAT:
		return n.floatLT(params)
	// case BOOL:
	// 	return n.boolLT(params)
	case STRING:
		return n.stringLT(params)
	}
	// This case should be unreachable if the parser is working correctly!
	panic(fmt.Sprintf("subexpression evaluated to impossible type %q", t))
}
