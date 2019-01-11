package rule

import "strconv"

func init() {
	Operators["add"] = func() Operator { return newExprAdd() }
	Operators["sub"] = func() Operator { return newExprSub() }
}

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

type exprSub struct {
	operator
}

func newExprSub() *exprSub {
	return &exprSub{
		operator: operator{
			contract: Contract{
				OpCode:     "sub",
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

// Sub creates an expression that takes at least two operands, which
// must evaluate to either Float64Value or Int64Value, and returns the
// value of subtracting the 2nd value from the first. Any values
// beyond the second one are subtracted from the preceding result
// until we reach the end of the values.
func Sub(vN ...Expr) Expr {
	e := newExprSub()
	e.consumeOperands(vN...)
	return e
}

func (n *exprSub) float64Sub(params Params) (*Value, error) {
	o0 := n.operands[0]
	v0, err := o0.Eval(params)
	if err != nil {
		return nil, err
	}

	f0, err := strconv.ParseFloat(v0.Data, 64)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.operands); i++ {
		o := n.operands[i]
		v, err := o.Eval(params)
		if err != nil {
			return nil, err
		}
		f, err := strconv.ParseFloat(v.Data, 64)
		if err != nil {
			return nil, err
		}
		f0 = f0 - f
	}
	return Float64Value(f0), nil
}

func (n *exprSub) int64Sub(params Params) (*Value, error) {
	o0 := n.operands[0]
	v0, err := o0.Eval(params)
	if err != nil {
		return nil, err
	}

	i0, err := strconv.ParseInt(v0.Data, 10, 64)
	if err != nil {
		return nil, err
	}

	for j := 1; j < len(n.operands); j++ {
		o := n.operands[j]
		v, err := o.Eval(params)
		if err != nil {
			return nil, err
		}
		i, err := strconv.ParseInt(v.Data, 10, 64)
		if err != nil {
			return nil, err
		}
		i0 = i0 - i
	}
	return Int64Value(i0), nil
}

// Eval makes exprSub comply with the Expr interface.
func (n *exprSub) Eval(params Params) (*Value, error) {
	// The ReturnType will be set to the concrete type that
	// matches all the arguments by homogenisation.
	if n.operator.Contract().ReturnType == FLOAT {
		return n.float64Sub(params)
	}
	return n.int64Sub(params)
}
