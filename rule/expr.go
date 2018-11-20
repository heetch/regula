package rule

import (
	"errors"
	"fmt"
	"go/token"
	"strconv"
)

var allOperators []string = []string{
	"not",
	"or",
	"and",
	"eq",
	"in",
}

// An Expr is a logical expression that can be evaluated to a value.
type Expr interface {
	Eval(Params) (*Value, error)
	Contract() Contract
}

// ComparableExpression is a logical expression that can be compared
// to another logical expression for equivalence, without evaluation.
type ComparableExpression interface {
	Same(ComparableExpression) bool
	GetKind() string
}

// IsOperator is a convenience function that identifies expressions that are operators.
func IsOperator(e Expr) bool {
	kind := e.GetKind()
	for _, op := range allOperators {
		if op == kind {
			return true
		}
	}
	return false
}

// A Params is a set of parameters passed on rule evaluation.
// It provides type safe methods to query params.
type Params interface {
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
	Keys() []string
	EncodeValue(key string) (string, error)
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

func (n *exprIn) Eval(params Params) (*Value, error) {
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

// Param is an expression used to select a parameter passed during evaluation and return its corresponding value.
type Param struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// Same compares the Param with a ComparableExpression to see if they
// are identical.  This is required by the ComparableExpression
// interface.
func (p *Param) Same(c ComparableExpression) bool {
	if p.Kind == c.GetKind() {
		p2, ok := c.(*Param)
		return ok && p.Type == p2.Type && p.Name == p2.Name
	}
	return false
}

// GetKind returns the Param's Kind in the way required by the
// ComparableExpression interface.
func (p *Param) GetKind() string {
	return p.Kind
}

// Contract returns the Contract of a param (which is simply a
// ReturnType that matches the param).  Thus Params implement the
// TypedExpression interface.
func (p *Param) Contract() Contract {
	switch p.Type {
	case "bool":
		return Contract{ReturnType: BOOLEAN}
	case "string":
		return Contract{ReturnType: STRING}
	case "int64":
		return Contract{ReturnType: INTEGER}
	case "float64":
		return Contract{ReturnType: FLOAT}
	}
	panic(fmt.Sprintf("invalid value type: %q", p.Type))
}

// StringParam creates a Param that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be a string. If not found it returns an error.
func StringParam(name string) *Param {
	return &Param{
		Kind: "param",
		Type: "string",
		Name: name,
	}
}

// BoolParam creates a Param that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be a boolean. If not found it returns an error.
func BoolParam(name string) *Param {
	return &Param{
		Kind: "param",
		Type: "bool",
		Name: name,
	}
}

// Int64Param creates a Param that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be an int64. If not found it returns an error.
func Int64Param(name string) *Param {
	return &Param{
		Kind: "param",
		Type: "int64",
		Name: name,
	}
}

// Float64Param creates a Param that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be a float64. If not found it returns an error.
func Float64Param(name string) *Param {
	return &Param{
		Kind: "param",
		Type: "float64",
		Name: name,
	}
}

// Eval extracts a value from the given parameters.
func (p *Param) Eval(params Params) (*Value, error) {
	if params == nil {
		return nil, errors.New("params is nil")
	}

	switch p.Type {
	case "string":
		v, err := params.GetString(p.Name)
		if err != nil {
			return nil, err
		}
		return StringValue(v), nil
	case "bool":
		v, err := params.GetBool(p.Name)
		if err != nil {
			return nil, err
		}
		return BoolValue(v), nil
	case "int64":
		v, err := params.GetInt64(p.Name)
		if err != nil {
			return nil, err
		}
		return Int64Value(v), nil
	case "float64":
		v, err := params.GetFloat64(p.Name)
		if err != nil {
			return nil, err
		}
		return Float64Value(v), nil
	}

	return nil, errors.New("unsupported param type")
}

// True creates an expression that always evaluates to true.
func True() Expr {
	return BoolValue(true)
}

// A Value is the result of the evaluation of an expression.
type Value struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Data string `json:"data"`
}

func newValue(typ, data string) *Value {
	return &Value{
		Kind: "value",
		Type: typ,
		Data: data,
	}
}

// Compares a Value with a ComparableExpression, without evaluating
// either.  This is required by the ComparableExpression interface.
func (v *Value) Same(c ComparableExpression) bool {
	if v.Kind == c.GetKind() {
		v2, ok := c.(*Value)
		return ok && v.Type == v2.Type && v.Data == v2.Data
	}
	return false
}

// GetKind returns the Value's kind, and is required by the ComparableExpression interface.
func (v *Value) GetKind() string {
	return v.Kind
}

// Contract returns the Contract of a value (which is simply a
// ReturnType that matches the value).  Thus Values implement the
// TypedExpression interface.
func (v *Value) Contract() Contract {
	switch v.Type {
	case "bool":
		return Contract{ReturnType: BOOLEAN}
	case "string":
		return Contract{ReturnType: STRING}
	case "int64":
		return Contract{ReturnType: INTEGER}
	case "float64":
		return Contract{ReturnType: FLOAT}
	}
	panic(fmt.Sprintf("invalid value type: %q", v.Type))
}

// BoolValue creates a bool type value.
func BoolValue(value bool) *Value {
	return newValue("bool", strconv.FormatBool(value))
}

// StringValue creates a string type value.
func StringValue(value string) *Value {
	return newValue("string", value)
}

// Int64Value creates an int64 type value.
func Int64Value(value int64) *Value {
	return newValue("int64", strconv.FormatInt(value, 10))
}

// Float64Value creates a float64 type value.
func Float64Value(value float64) *Value {
	return newValue("float64", strconv.FormatFloat(value, 'f', 6, 64))
}

// Eval evaluates the value to itself.
func (v *Value) Eval(Params) (*Value, error) {
	return v, nil
}

func (v *Value) compare(op token.Token, other *Value) bool {
	if op != token.EQL {
		return false
	}

	return *v == *other
}

// Equal reports whether v and other represent the same value.
func (v *Value) Equal(other *Value) bool {
	return v.compare(token.EQL, other)
}

// Operander is an interface for managing the operands of an
// Expr that is an operation.
type Operander interface {
	// Operands returns all of the operands currently held by the Operander.
	Operands() []Expr
	// PushExpr adds an Expr as an operand.
	PushExpr(e Expr) error
	// Finalise indicates that we are done pushing Expr's to the Operander.  This allows for arity checking.
	Finalise() error
}

func walk(expr Expr, fn func(Expr) error) error {
	err := fn(expr)
	if err != nil {
		return err
	}

	if o, ok := expr.(Operander); ok {
		ops := o.Operands()
		for _, op := range ops {
			err := walk(op, fn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
