package rule

import "strconv"

// exprIntToFloat is an operation that converts an Int64 type to a Float64 type.
type exprIntToFloat struct {
	operator
}

// newExprIntToFloat returns a pointer to a exprIntToFloat operator, initialised with a Contract.
func newExprIntToFloat() *exprIntToFloat {
	return &exprIntToFloat{
		operator: operator{
			contract: Contract{
				OpCode:     "intToFloat",
				ReturnType: FLOAT,
				Terms: []Term{
					{
						Cardinality: ONE,
						Type:        INTEGER,
					},
				},
			},
		},
	}
}

// IntToFloat is an operator that converts as Float64 into an Int64 type.
func IntToFloat(vN ...Expr) Expr {
	e := newExprIntToFloat()
	e.consumeOperands(vN...)
	return e
}

// Eval will convert the operand provided to exprIntToFloat to a Float64Value.  Eval makes exprIntToFloat implement the Expr interface.
func (n *exprIntToFloat) Eval(params Params) (*Value, error) {
	op := n.operands[0]
	val, err := op.Eval(params)
	if err != nil {
		return nil, err
	}
	iVal, err := strconv.ParseInt(val.Data, 10, 64)
	if err != nil {
		return nil, err
	}
	return Float64Value(float64(iVal)), nil
}
