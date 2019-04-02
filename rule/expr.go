package rule

import (
	"errors"
	"go/token"
	"strconv"
)

// An Expr is a logical expression that can be evaluated to a value.
type Expr interface {
	Eval(Params) (*Value, error)
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

// Not creates an expression that evaluates the given operand e and returns its opposite.
// e must evaluate to a boolean.
func Not(e Expr) Expr {
	return &exprNot{
		operator: operator{
			kind:     "not",
			operands: []Expr{e},
		},
	}
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

// Or creates an expression that takes at least two operands and evaluates to true if one of the operands evaluates to true.
// All the given operands must evaluate to a boolean.
func Or(v1, v2 Expr, vN ...Expr) Expr {
	return &exprOr{
		operator: operator{
			kind:     "or",
			operands: append([]Expr{v1, v2}, vN...),
		},
	}
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

// And creates an expression that takes at least two operands and evaluates to true if all the operands evaluate to true.
// All the given operands must evaluate to a boolean.
func And(v1, v2 Expr, vN ...Expr) Expr {
	return &exprAnd{
		operator: operator{
			kind:     "and",
			operands: append([]Expr{v1, v2}, vN...),
		},
	}
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

// Eq creates an expression that takes at least two operands and evaluates to true if all the operands are equal.
func Eq(v1, v2 Expr, vN ...Expr) Expr {
	return &exprEq{
		operator: operator{
			kind:     "eq",
			operands: append([]Expr{v1, v2}, vN...),
		},
	}
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

// In creates an expression that takes at least two operands and evaluates to true if the first one is equal to one of the others.
func In(v, e1 Expr, eN ...Expr) Expr {
	return &exprIn{
		operator: operator{
			kind:     "in",
			operands: append([]Expr{v, e1}, eN...),
		},
	}
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

type exprGT struct {
	operator
}

// Gt creates an expression that takes at least two operands and
// evaluates to true if each successive operand has a higher value than
// the next.
func Gt(v1, v2 Expr, vN ...Expr) Expr {
	return &exprGT{
		operator: operator{
			kind:     "gt",
			operands: append([]Expr{v1, v2}, vN...),
		},
	}
}

func (n *exprGT) Eval(params Params) (*Value, error) {
	if len(n.operands) < 2 {
		return nil, errors.New("invalid number of operands in Gt func")
	}

	vA, err := n.operands[0].Eval(params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.operands); i++ {
		vB, err := n.operands[i].Eval(params)
		if err != nil {
			return nil, err
		}

		if !vA.Gt(vB) {
			return BoolValue(false), nil
		}
	}

	return BoolValue(true), nil
}

// Param is an expression used to select a parameter passed during evaluation and return its corresponding value.
type Param struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Name string `json:"name"`
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

// Gt reports whether v is greater than other.
func (v *Value) Gt(other *Value) bool {
	switch v.Type {
	case "bool":
		v1, _ := strconv.ParseBool(v.Data)
		if !v1 {
			// If v1 is False then it's not greater than v2, and we can be done already.
			return false
		}

		v2, _ := strconv.ParseBool(other.Data)
		if v2 {
			// If v2 is True then v1 can't be greater than it..
			return false
		}
		return true
	case "string":
		if v.Data <= other.Data {
			return false
		}
		return true
	case "int64":
		v1, _ := strconv.ParseInt(v.Data, 10, 64)
		v2, _ := strconv.ParseInt(other.Data, 10, 64)
		if v1 <= v2 {
			return false
		}
		return true
	case "float64":
		v1, _ := strconv.ParseFloat(v.Data, 64)
		v2, _ := strconv.ParseFloat(other.Data, 64)
		if v1 <= v2 {
			return false
		}
		return true
	}
	return false
}

type operander interface {
	Operands() []Expr
}

func walk(expr Expr, fn func(Expr) error) error {
	err := fn(expr)
	if err != nil {
		return err
	}

	if o, ok := expr.(operander); ok {
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
