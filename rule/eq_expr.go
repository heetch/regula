package rule

import (
	"errors"

	"github.com/heetch/regula/param"
)

func init() {
	Operators["eq"] = func() Operator { return newExprEq() }

}

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

func (n *exprEq) Eval(params param.Params) (*Value, error) {
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
