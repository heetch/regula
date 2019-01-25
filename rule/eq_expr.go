package rule

import (
	"errors"
	"fmt"
)

func init() {
	Operators["eq"] = func() Operator { return newExprEq() }
	Operators["lt"] = func() Operator { return newExprLT() }
	Operators["gt"] = func() Operator { return newExprGT() }
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
	return eq(n.operands, params)
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

func (n *exprLT) Eval(params Params) (*Value, error) {
	return lt(n.operands, params)
}

/////////////////
// LTE operator //
/////////////////

type exprLTE struct {
	operator
}

func newExprLTE() *exprLTE {
	return &exprLTE{
		operator: operator{
			contract: Contract{
				OpCode:     "lte",
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

// LTE creates an expression that takes at least two operands and
// evaluates to true if each successive operand has a lower or equal
// value compared to the next.
func LTE(vN ...Expr) Expr {
	e := newExprLTE()
	e.consumeOperands(vN...)
	return e
}

func (n *exprLTE) Eval(params Params) (*Value, error) {
	return lte(n.operands, params)
}

/////////////////
// GT operator //
/////////////////

type exprGT struct {
	operator
}

func newExprGT() *exprGT {
	return &exprGT{
		operator: operator{
			contract: Contract{
				OpCode:     "gt",
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

// GT creates an expression that takes at least two operands and
// evaluates to true if each successive operand has a higher value than
// the next.
func GT(vN ...Expr) Expr {
	e := newExprGT()
	e.consumeOperands(vN...)
	return e
}

func (n *exprGT) Eval(params Params) (*Value, error) {
	return gt(n.operands, params)
}

/////////////////////////////////
// Underlying type comparators //
/////////////////////////////////

// perform an equality check on a sequence of Exprs
func eq(operands []Expr, params Params) (*Value, error) {
	if len(operands) < 2 {
		return nil, errors.New("invalid number of operands in Eq func")
	}

	opA := operands[0]
	vA, err := opA.Eval(params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(operands); i++ {
		vB, err := operands[i].Eval(params)
		if err != nil {
			return nil, err
		}

		if !vA.Equal(vB) {
			return BoolValue(false), nil
		}
	}

	return BoolValue(true), nil
}

// perform a greater-than comparison on a sequence of expressions
func gt(operands []Expr, params Params) (*Value, error) {
	// Because of homogenisation during Parsing we know that all
	// operands have the same type.
	t := operands[0].Contract().ReturnType
	switch t {
	case INTEGER:
		return integerGT(operands, params)
	case FLOAT:
		return floatGT(operands, params)
	case BOOLEAN:
		return boolGT(operands, params)
	case STRING:
		return stringGT(operands, params)
	}
	// This case should be unreachable if the parser is working correctly!
	panic(fmt.Sprintf("subexpression evaluated to impossible type %q", t))
}

// perform a greater-than comparison on a sequence of integers
func integerGT(operands []Expr, params Params) (*Value, error) {
	var i0, i1 int64
	var err error

	i0, err = exprToInt64(operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(operands); j++ {
		i1, err = exprToInt64(operands[j], params)
		if err != nil {
			return nil, err
		}
		if i0 <= i1 {
			return BoolValue(false), nil
		}
		i0 = i1
	}
	return BoolValue(true), nil
}

// perform a greater-than comparison on a sequence of floats
func floatGT(operands []Expr, params Params) (*Value, error) {
	var f0, f1 float64
	var err error

	f0, err = exprToFloat64(operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(operands); j++ {
		f1, err = exprToFloat64(operands[j], params)
		if err != nil {
			return nil, err
		}
		if f0 <= f1 {
			return BoolValue(false), nil
		}
		f0 = f1
	}
	return BoolValue(true), nil
}

// perform a greater-than comparison on a sequence of bools
func boolGT(operands []Expr, params Params) (*Value, error) {
	var b0, b1 bool
	var err error

	if len(operands) > 2 {
		// We can't have greater than 2 operands and maintain
		// an inequality with a binary choice.
		return BoolValue(false), nil
	}

	b0, err = exprToBool(operands[0], params)
	if err != nil {
		return nil, err
	}
	if !b0 {
		// If b0 is False then it's not greater than b1, and we can be done already.
		return BoolValue(false), nil
	}
	b1, err = exprToBool(operands[1], params)
	if err != nil {
		return nil, err
	}
	if b1 {
		// If b1 is True then b0 can't be greater than it..
		return BoolValue(false), nil
	}
	return BoolValue(true), nil
}

// perform a greater-than comparison on a sequence of strings
func stringGT(operands []Expr, params Params) (*Value, error) {
	var s0, s1 string
	var err error

	s0, err = exprToString(operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(operands); j++ {
		s1, err = exprToString(operands[j], params)
		if err != nil {
			return nil, err
		}
		if s0 <= s1 {
			return BoolValue(false), nil
		}
		s0 = s1
	}
	return BoolValue(true), nil
}

// perform a less-than comparison on a sequence of Exprs
func lt(operands []Expr, params Params) (*Value, error) {
	// Because of homogenisation during Parsing we know that all
	// operands have the same type.
	t := operands[0].Contract().ReturnType
	switch t {
	case INTEGER:
		return integerLT(operands, params)
	case FLOAT:
		return floatLT(operands, params)
	case BOOLEAN:
		return boolLT(operands, params)
	case STRING:
		return stringLT(operands, params)
	}
	// This case should be unreachable if the parser is working correctly!
	panic(fmt.Sprintf("subexpression evaluated to impossible type %q", t))
}

// perform a less-than comparison on a sequence of integers
func integerLT(operands []Expr, params Params) (*Value, error) {
	var i0, i1 int64
	var err error

	i0, err = exprToInt64(operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(operands); j++ {
		i1, err = exprToInt64(operands[j], params)
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
func floatLT(operands []Expr, params Params) (*Value, error) {
	var f0, f1 float64
	var err error

	f0, err = exprToFloat64(operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(operands); j++ {
		f1, err = exprToFloat64(operands[j], params)
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

// perform a less-than comparison on a sequence of bools
func boolLT(operands []Expr, params Params) (*Value, error) {
	var b0, b1 bool
	var err error

	if len(operands) > 2 {
		// We can't have greater than 2 operands and maintain
		// an inequality with a binary choice.
		return BoolValue(false), nil
	}

	b0, err = exprToBool(operands[0], params)
	if err != nil {
		return nil, err
	}
	if b0 {
		// If b0 is True then it's not less than b1, and we can be done already.
		return BoolValue(false), nil
	}
	b1, err = exprToBool(operands[1], params)
	if err != nil {
		return nil, err
	}
	if !b1 {
		// If b1 is False then b0 can't be less than it..
		return BoolValue(false), nil
	}
	return BoolValue(true), nil
}

// perform a less-than comparison on a sequence of strings
func stringLT(operands []Expr, params Params) (*Value, error) {
	var s0, s1 string
	var err error

	s0, err = exprToString(operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(operands); j++ {
		s1, err = exprToString(operands[j], params)
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

// perform a less-than-or-equal comparison on a sequence of Exprs
func lte(operands []Expr, params Params) (*Value, error) {
	// Because of homogenisation during Parsing we know that all
	// operands have the same type.
	t := operands[0].Contract().ReturnType
	switch t {
	case INTEGER:
		return integerLTE(operands, params)
	case FLOAT:
		return floatLTE(operands, params)
	case BOOLEAN:
		return boolLTE(operands, params)
	case STRING:
		return stringLTE(operands, params)
	}
	// This case should be unreachable if the parser is working correctly!
	panic(fmt.Sprintf("subexpression evaluated to impossible type %q", t))

	// res1, err := lt(operands, params)
	// if err != nil {
	// 	return nil, err
	// }
	// res2, err := eq(operands, params)
	// if err != nil {
	// 	return nil, err
	// }
	// b1, err := exprToBool(res1, params)
	// if err != nil {
	// 	return nil, err
	// }
	// b2, err := exprToBool(res2, params)
	// if err != nil {
	// 	return nil, err
	// }
	// return BoolValue(b1 || b2), nil
}

// perform a less-than comparison on a sequence of integers
func integerLTE(operands []Expr, params Params) (*Value, error) {
	var i0, i1 int64
	var err error

	i0, err = exprToInt64(operands[0], params)
	if err != nil {
		return nil, err
	}

	for j := 1; j < len(operands); j++ {
		i1, err = exprToInt64(operands[j], params)
		if err != nil {
			return nil, err
		}
		if !(i0 <= i1) {
			return BoolValue(false), nil
		}
		i0 = i1
	}
	return BoolValue(true), nil
}

// perform a less-than comparison on a sequence of floats
func floatLTE(operands []Expr, params Params) (*Value, error) {
	var f0, f1 float64
	var err error

	f0, err = exprToFloat64(operands[0], params)
	if err != nil {
		return nil, err
	}
	for j := 1; j < len(operands); j++ {
		f1, err = exprToFloat64(operands[j], params)
		if err != nil {
			return nil, err
		}
		if !(f0 <= f1) {
			return BoolValue(false), nil
		}
		f0 = f1
	}
	return BoolValue(true), nil
}

// perform a less-than comparison on a sequence of bools
func boolLTE(operands []Expr, params Params) (*Value, error) {
	var b0, b1 bool
	var err error

	for i := 1; i < len(operands); i++ {
		b0, err = exprToBool(operands[i-1], params)
		if err != nil {
			return nil, err
		}
		b1, err = exprToBool(operands[i], params)
		if err != nil {
			return nil, err
		}
		if b0 && !b1 {
			return BoolValue(false), nil
		}
	}
	return BoolValue(true), nil
}

// perform a less-than comparison on a sequence of strings
func stringLTE(operands []Expr, params Params) (*Value, error) {
	var s0, s1 string
	var err error

	for i := 1; i < len(operands); i++ {
		s0, err = exprToString(operands[i-1], params)
		if err != nil {
			return nil, err
		}
		s1, err = exprToString(operands[i], params)
		if err != nil {
			return nil, err
		}
		if !(s0 <= s1) {
			return BoolValue(false), nil
		}
	}
	return BoolValue(true), nil
}
