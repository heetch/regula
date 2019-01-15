package rule

import (
	"errors"
	"fmt"
	"go/token"
	"strconv"

	"github.com/heetch/regula/param"
)

// An Expr is a logical expression that can be evaluated to a value.
type Expr interface {
	Eval(param.Params) (*Value, error)
	Contract() Contract
}

// ComparableExpression is a logical expression that can be compared
// to another logical expression for equivalence, without evaluation.
type ComparableExpression interface {
	Same(ComparableExpression) bool
	GetKind() string
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
func (p *Param) Eval(params param.Params) (*Value, error) {
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
func (v *Value) Eval(param.Params) (*Value, error) {
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

// exprToInt64 returns the go-native int64 value of an expression
// evaluated with params.
func exprToInt64(e Expr, params param.Params) (int64, error) {
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
func exprToFloat64(e Expr, params param.Params) (float64, error) {
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

// exprToBool returns the go-native bool value of an expression
// evaluated with params.
func exprToBool(e Expr, params param.Params) (bool, error) {
	v, err := e.Eval(params)
	if err != nil {
		return false, err
	}
	b, err := strconv.ParseBool(v.Data)
	if err != nil {
		return false, err
	}
	return b, nil
}
