package rule

import (
	"strconv"
)

func init() {
	Operators["intToFloat"] = func() Operator { return newExprIntToFloat() }
	Operators["floatToInt"] = func() Operator { return newExprFloatToInt() }
}

/////////////////////////
// IntToFloat Operator //
/////////////////////////

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
func IntToFloat(v Expr) Expr {
	e := newExprIntToFloat()
	e.consumeOperands(v)
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

/////////////////////////
// FloatToInt Operator //
/////////////////////////

// exprFloatToInt is an operation that converts a Float64 type to an Int64 type.
type exprFloatToInt struct {
	operator
}

// newExprFloatToInt returns a pointer to a exprFloatToInt operator, initialised with a Contract.
func newExprFloatToInt() *exprFloatToInt {
	return &exprFloatToInt{
		operator: operator{
			contract: Contract{
				OpCode:     "floatToInt",
				ReturnType: INTEGER,
				Terms: []Term{
					{
						Cardinality: ONE,
						Type:        FLOAT,
					},
				},
			},
		},
	}
}

// FloatToInt is an operator that converts a Float64 into an Int64 type.
func FloatToInt(v Expr) Expr {
	e := newExprFloatToInt()
	e.consumeOperands(v)
	return e
}

// Eval will convert the operand provided to exprFloatToInt to an Int64Value.  Eval makes exprFloatToInt implement the Expr interface.
func (n *exprFloatToInt) Eval(params Params) (*Value, error) {
	op := n.operands[0]
	val, err := op.Eval(params)
	if err != nil {
		return nil, err
	}
	// Handily this will just give us the int we want ;-)
	fVal, err := strconv.ParseFloat(val.Data, 64)
	if err != nil {
		return nil, err
	}
	return Int64Value(int64(fVal)), nil
}
