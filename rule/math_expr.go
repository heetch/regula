package rule

import "strconv"

type exprAdd struct {
	operator
}

func newExprAdd() *exprAdd {
	return &exprAdd{
		operator: operator{
			contract: Contract{
				OpCode:     "add",
				ReturnType: NUMBER,
				Terms: []Term{
					{
						Type:        NUMBER,
						Cardinality: MANY,
						Min:         2,
					},
				},
			},
		},
	}
}

// Add creates an expression that takes at least two operands, which must evaluate to either Float64Value or Int64Value, and returns their sum.
func Add(vN ...Expr) Expr {
	e := newExprAdd()
	e.consumeOperands(vN...)
	return e
}

func (n *exprAdd) float64Add(params Params) (*Value, error) {
	var sum float64

	for _, o := range n.operands {
		v, err := o.Eval(params)
		if err != nil {
			return nil, err
		}
		f, err := strconv.ParseFloat(v.Data, 64)
		if err != nil {
			return nil, err
		}
		sum += f
	}
	return Float64Value(sum), nil
}

func (n *exprAdd) int64Add(params Params) (*Value, error) {
	var sum int64

	for _, o := range n.operands {
		v, err := o.Eval(params)
		if err != nil {
			return nil, err
		}
		i, err := strconv.ParseInt(v.Data, 10, 64)
		if err != nil {
			return nil, err
		}
		sum += i
	}
	return Int64Value(sum), nil
}

// Eval makes exprAdd comply with the Expr interface.
func (n *exprAdd) Eval(params Params) (*Value, error) {
	// The ReturnType will be set to the concrete type that
	// matches all the arguments by homogenisation.
	if n.operator.Contract().ReturnType == FLOAT {
		return n.float64Add(params)
	}
	return n.int64Add(params)
}
