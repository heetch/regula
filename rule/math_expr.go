package rule

import "strconv"

func init() {
	Operators["add"] = func() Operator { return newExprAdd() }
	Operators["sub"] = func() Operator { return newExprSub() }
	Operators["mult"] = func() Operator { return newExprMult() }
	Operators["div"] = func() Operator { return newExprDiv() }
	Operators["mod"] = func() Operator { return newExprMod() }
}

//////////////////
// Add operator //
//////////////////

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

// Add creates an expression that takes at least two operands, which
// must evaluate to either Float64Value or Int64Value, and returns
// their sum.
func Add(vN ...Expr) Expr {
	e := newExprAdd()
	e.consumeOperands(vN...)
	return e
}

func (n *exprAdd) float64Add(params Params) (*Value, error) {
	var sum float64

	for _, o := range n.operands {
		f, err := exprToFloat64(o, params)
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
		i, err := exprToInt64(o, params)
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

//////////////////
// Sub operator //
//////////////////
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
	f0, err := exprToFloat64(n.operands[0], params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.operands); i++ {
		f, err := exprToFloat64(n.operands[i], params)
		if err != nil {
			return nil, err
		}
		f0 = f0 - f
	}
	return Float64Value(f0), nil
}

func (n *exprSub) int64Sub(params Params) (*Value, error) {
	i0, err := exprToInt64(n.operands[0], params)
	if err != nil {
		return nil, err
	}

	for j := 1; j < len(n.operands); j++ {
		i, err := exprToInt64(n.operands[j], params)
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

///////////////////
// Mult Operator //
///////////////////
type exprMult struct {
	operator
}

func newExprMult() *exprMult {
	return &exprMult{
		operator: operator{
			contract: Contract{
				OpCode:     "mult",
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

// Mult creates an expression that takes at least two operands, which
// must evaluate to either Float64Value or Int64Value, and returns the
// product of all operands.
func Mult(vN ...Expr) Expr {
	e := newExprMult()
	e.consumeOperands(vN...)
	return e
}

// Perform multiplication of Float64Value types.
func (n *exprMult) float64Mult(params Params) (*Value, error) {
	var product float64 = 1.0
	for _, o := range n.operands {
		f, err := exprToFloat64(o, params)
		if err != nil {
			return nil, err
		}
		product = product * f
	}
	return Float64Value(product), nil
}

// Perform multiplication of Int64Value types.
func (n *exprMult) int64Mult(params Params) (*Value, error) {
	var product int64 = 1
	for _, o := range n.operands {
		i, err := exprToInt64(o, params)
		if err != nil {
			return nil, err
		}
		product = product * i
	}
	return Int64Value(product), nil
}

// Eval makes exprMult comply with the Expr interface.
func (n *exprMult) Eval(params Params) (*Value, error) {
	// The ReturnType will be set to the concrete type that
	// matches all the arguments by homogenisation.
	if n.operator.Contract().ReturnType == FLOAT {
		return n.float64Mult(params)
	}
	return n.int64Mult(params)
}

///////////////////
// Div Operator //
///////////////////
type exprDiv struct {
	operator
}

func newExprDiv() *exprDiv {
	return &exprDiv{
		operator: operator{
			contract: Contract{
				OpCode:     "div",
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

// Div creates an expression that takes at least two operands, which
// must evaluate to either Float64Value or Int64Value, and returns the
// value of the first operand successively divided by the operands
// that follow.
func Div(vN ...Expr) Expr {
	e := newExprDiv()
	e.consumeOperands(vN...)
	return e
}

// Perform division of Float64Value types.
func (n *exprDiv) float64Div(params Params) (*Value, error) {
	quotient, err := exprToFloat64(n.operands[0], params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.operands); i++ {
		f, err := exprToFloat64(n.operands[i], params)
		if err != nil {
			return nil, err
		}
		quotient = quotient / f
	}
	return Float64Value(quotient), nil
}

// Perform division of Int64Value types.
func (n *exprDiv) int64Div(params Params) (*Value, error) {
	quotient, err := exprToInt64(n.operands[0], params)
	if err != nil {
		return nil, err
	}

	for j := 1; j < len(n.operands); j++ {
		i, err := exprToInt64(n.operands[j], params)
		if err != nil {
			return nil, err
		}
		quotient = quotient / i
	}
	return Int64Value(quotient), nil
}

// Eval makes exprDiv comply with the Expr interface.
func (n *exprDiv) Eval(params Params) (*Value, error) {
	// The ReturnType will be set to the concrete type that
	// matches all the arguments by homogenisation.
	if n.operator.Contract().ReturnType == FLOAT {
		return n.float64Div(params)
	}
	return n.int64Div(params)
}

///////////////////
// Mod Operator //
///////////////////
type exprMod struct {
	operator
}

func newExprMod() *exprMod {
	return &exprMod{
		operator: operator{
			contract: Contract{
				OpCode:     "mod",
				ReturnType: INTEGER,
				Terms: []Term{
					{
						Type:        INTEGER,
						Cardinality: ONE,
					},
					{
						Type:        INTEGER,
						Cardinality: ONE,
					},
				},
			},
		},
	}
}

// Mod creates an expression that takes exactly two operands, which
// must evaluate to Int64Value, and returns the remainder of Euclidean
// division of the first integer value by the second.
func Mod(v0, v1 Expr) Expr {
	e := newExprMod()
	e.pushExprOrPanic(v0)
	e.pushExprOrPanic(v1)
	e.finaliseOrPanic()
	return e
}

// Eval makes exprMod comply with the Expr interface.
func (n *exprMod) Eval(params Params) (*Value, error) {
	dividend, err := exprToInt64(n.operands[0], params)
	if err != nil {
		return nil, err
	}
	divisor, err := exprToInt64(n.operands[1], params)
	if err != nil {
		return nil, err
	}
	return Int64Value(dividend % divisor), nil
}

///////////////////////
// Utility functions //
///////////////////////

// exprToInt64 returns the go-native int64 value of an expression
// evaluated with params.
func exprToInt64(e Expr, params Params) (int64, error) {
	v, err := e.Eval(params)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(v.Data, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, err
}

// exprToFloat64 returns the go-native float64 value of an expression
// evaluated with params.
func exprToFloat64(e Expr, params Params) (float64, error) {
	v, err := e.Eval(params)
	if err != nil {
		return 0.0, err
	}
	f, err := strconv.ParseFloat(v.Data, 64)
	if err != nil {
		return 0.0, err
	}
	return f, nil
}
