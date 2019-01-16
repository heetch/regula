package rule

import (
	"errors"
)

func init() {
	Operators["and"] = func() Operator { return newExprAnd() }
	Operators["not"] = func() Operator { return newExprNot() }
	Operators["or"] = func() Operator { return newExprOr() }
}

type exprNot struct {
	operator
}

func newExprNot() *exprNot {
	return &exprNot{
		operator: operator{
			contract: Contract{
				OpCode:     "not",
				ReturnType: BOOLEAN,
				Terms:      []Term{{Type: BOOLEAN, Cardinality: ONE}},
			},
		},
	}
}

// Not creates an expression that evaluates the given operand e and returns its opposite.
// e must evaluate to a boolean.
func Not(e Expr) Expr {
	expr := newExprNot()
	expr.consumeOperands(e)
	return expr
}

func (n *exprNot) Eval(params Params) (*Value, error) {
	if len(n.operands) < 1 {
		return nil, errors.New("invalid number of operands in Not func")
	}

	op := n.operands[0]
	v, err := op.Eval(params)
	if err != nil {
		return nil, err
	}

	if v.Type != "bool" {
		return nil, errors.New("invalid operand type for Not func")
	}

	if v.Equal(BoolValue(true)) {
		return BoolValue(false), nil
	}

	return BoolValue(true), nil
}

type exprOr struct {
	operator
}

func newExprOr() *exprOr {
	return &exprOr{
		operator: operator{
			contract: Contract{
				OpCode:     "or",
				ReturnType: BOOLEAN,
				Terms:      []Term{{Type: BOOLEAN, Cardinality: MANY, Min: 2}},
			},
		},
	}
}

// Or creates an expression that takes at least two operands and evaluates to true if one of the operands evaluates to true.
// All the given operands must evaluate to a boolean.
func Or(vN ...Expr) Expr {
	e := newExprOr()
	e.consumeOperands(vN...)
	return e
}

func (n *exprOr) Eval(params Params) (*Value, error) {
	if len(n.operands) < 2 {
		return nil, errors.New("invalid number of operands in Or func")
	}

	opA := n.operands[0]
	vA, err := opA.Eval(params)
	if err != nil {
		return nil, err
	}
	if vA.Type != "bool" {
		return nil, errors.New("invalid operand type for Or func")
	}

	if vA.Equal(BoolValue(true)) {
		return vA, nil
	}

	for i := 1; i < len(n.operands); i++ {
		vB, err := n.operands[i].Eval(params)
		if err != nil {
			return nil, err
		}
		if vB.Type != "bool" {
			return nil, errors.New("invalid operand type for Or func")
		}

		if vB.Equal(BoolValue(true)) {
			return vB, nil
		}
	}

	return BoolValue(false), nil
}

type exprAnd struct {
	operator
}

func newExprAnd() *exprAnd {
	return &exprAnd{
		operator: operator{
			contract: Contract{
				OpCode:     "and",
				ReturnType: BOOLEAN,
				Terms: []Term{
					{
						Type:        BOOLEAN,
						Cardinality: MANY,
						Min:         2,
					},
				},
			},
		},
	}
}

// And creates an expression that takes at least two operands and evaluates to true if all the operands evaluate to true.
// All the given operands must evaluate to a boolean.
func And(vN ...Expr) Expr {
	e := newExprAnd()
	e.consumeOperands(vN...)
	return e
}

func (n *exprAnd) Eval(params Params) (*Value, error) {
	if len(n.operands) < 2 {
		return nil, errors.New("invalid number of operands in And func")
	}

	opA := n.operands[0]
	vA, err := opA.Eval(params)
	if err != nil {
		return nil, err
	}
	if vA.Type != "bool" {
		return nil, errors.New("invalid operand type for And func")
	}

	if vA.Equal(BoolValue(false)) {
		return vA, nil
	}

	for i := 1; i < len(n.operands); i++ {
		vB, err := n.operands[i].Eval(params)
		if err != nil {
			return nil, err
		}
		if vB.Type != "bool" {
			return nil, errors.New("invalid operand type for And func")
		}

		if vB.Equal(BoolValue(false)) {
			return vB, nil
		}
	}

	return BoolValue(true), nil
}
