package rule

import (
	"errors"

	"github.com/heetch/regula/param"
)

func init() {
	Operators["in"] = func() Operator { return newExprIn() }
}

type exprIn struct {
	operator
}

func newExprIn() *exprIn {
	return &exprIn{
		operator: operator{
			contract: Contract{
				OpCode:     "in",
				ReturnType: BOOLEAN,
				Terms: []Term{
					{
						Type:        ANY,
						Cardinality: ONE,
					},
					{
						Type:        ANY,
						Cardinality: MANY,
						Min:         1,
					},
				},
			},
		},
	}
}

// In creates an expression that takes at least two operands and evaluates to true if the first one is equal to one of the others.
func In(vN ...Expr) Expr {
	e := newExprIn()
	e.consumeOperands(vN...)
	return e
}

func (n *exprIn) Eval(params param.Params) (*Value, error) {
	if len(n.operands) < 2 {
		return nil, errors.New("invalid number of operands in In func")
	}

	toFind := n.operands[0]
	vA, err := toFind.Eval(params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.operands); i++ {
		vB, err := n.operands[i].Eval(params)
		if err != nil {
			return nil, err
		}

		if vA.Equal(vB) {
			return BoolValue(true), nil
		}
	}

	return BoolValue(false), nil
}
