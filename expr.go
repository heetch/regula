package regula

import (
	"encoding/json"
	"errors"
	"go/token"
	"strconv"

	"github.com/tidwall/gjson"
)

// An Expr is a logical expression that can be evaluated to a value.
type Expr interface {
	Eval(ParamGetter) (*Value, error)
}

func parseExpr(kind string, data []byte) (Expr, error) {
	var e Expr
	var err error

	switch kind {
	case "eq":
		var eq exprEq
		e = &eq
		err = eq.UnmarshalJSON(data)
	case "in":
		var in exprIn
		e = &in
		err = in.UnmarshalJSON(data)
	case "not":
		var not exprNot
		e = &not
		err = not.UnmarshalJSON(data)
	case "and":
		var and exprAnd
		e = &and
		err = and.UnmarshalJSON(data)
	case "or":
		var or exprOr
		e = &or
		err = or.UnmarshalJSON(data)
	case "value":
		var v Value
		e = &v
		err = json.Unmarshal(data, &v)
	case "param":
		var v exprParam
		e = &v
		err = json.Unmarshal(data, &v)
	default:
		err = errors.New("unknown expression kind")
	}

	return e, err
}

type operands struct {
	Ops   []json.RawMessage `json:"operands"`
	Exprs []Expr
}

func (o *operands) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &o.Ops)
	if err != nil {
		return err
	}

	for _, op := range o.Ops {
		r := gjson.Get(string(op), "kind")
		n, err := parseExpr(r.Str, []byte(op))
		if err != nil {
			return err
		}

		o.Exprs = append(o.Exprs, n)
	}

	return nil
}

type nodeOps struct {
	Kind     string   `json:"kind"`
	Operands operands `json:"operands"`
}

type exprNot struct {
	Kind     string `json:"kind"`
	Operands []Expr `json:"operands"`
}

// Not creates an expression that evaluates the given operand e and returns its opposite.
// e must evaluate to a boolean.
func Not(e Expr) Expr {
	return &exprNot{
		Kind:     "not",
		Operands: []Expr{e},
	}
}

func (n *exprNot) Eval(params ParamGetter) (*Value, error) {
	if len(n.Operands) < 1 {
		return nil, errors.New("invalid number of operands in Not func")
	}

	op := n.Operands[0]
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

func (n *exprNot) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Exprs) < 1 {
		return errors.New("invalid number of operands in Not func")
	}

	n.Operands = node.Operands.Exprs
	n.Kind = "not"

	return nil
}

type exprOr struct {
	Kind     string `json:"kind"`
	Operands []Expr `json:"operands"`
}

// Or creates an expression that takes at least two operands and evaluates to true if one of the operands evaluates to true.
// All the given operands must evaluate to a boolean.
func Or(v1, v2 Expr, vN ...Expr) Expr {
	return &exprOr{
		Kind:     "or",
		Operands: append([]Expr{v1, v2}, vN...),
	}
}

func (n *exprOr) Eval(params ParamGetter) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in Or func")
	}

	opA := n.Operands[0]
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

	for i := 1; i < len(n.Operands); i++ {
		vB, err := n.Operands[i].Eval(params)
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

func (n *exprOr) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Exprs) < 2 {
		return errors.New("invalid number of operands in Or func")
	}

	or := Or(node.Operands.Exprs[0], node.Operands.Exprs[1], node.Operands.Exprs[2:]...)
	*n = *(or.(*exprOr))
	return nil
}

type exprAnd struct {
	Kind     string `json:"kind"`
	Operands []Expr `json:"operands"`
}

// And creates an expression that takes at least two operands and evaluates to true if all the operands evaluate to true.
// All the given operands must evaluate to a boolean.
func And(v1, v2 Expr, vN ...Expr) Expr {
	return &exprAnd{
		Kind:     "and",
		Operands: append([]Expr{v1, v2}, vN...),
	}
}

func (n *exprAnd) Eval(params ParamGetter) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in And func")
	}

	opA := n.Operands[0]
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

	for i := 1; i < len(n.Operands); i++ {
		vB, err := n.Operands[i].Eval(params)
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

func (n *exprAnd) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Exprs) < 2 {
		return errors.New("invalid number of operands in And func")
	}

	and := And(node.Operands.Exprs[0], node.Operands.Exprs[1], node.Operands.Exprs[2:]...)
	*n = *(and.(*exprAnd))
	return nil
}

type exprEq struct {
	Kind     string `json:"kind"`
	Operands []Expr `json:"operands"`
}

// Eq creates an expression that takes at least two operands and evaluates to true if all the operands are equal.
func Eq(v1, v2 Expr, vN ...Expr) Expr {
	return &exprEq{
		Kind:     "eq",
		Operands: append([]Expr{v1, v2}, vN...),
	}
}

func (n *exprEq) Eval(params ParamGetter) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in Eq func")
	}

	opA := n.Operands[0]
	vA, err := opA.Eval(params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.Operands); i++ {
		vB, err := n.Operands[i].Eval(params)
		if err != nil {
			return nil, err
		}

		if !vA.Equal(vB) {
			return BoolValue(false), nil
		}
	}

	return BoolValue(true), nil
}

func (n *exprEq) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Exprs) < 2 {
		return errors.New("invalid number of operands in Eq func")
	}

	eq := Eq(node.Operands.Exprs[0], node.Operands.Exprs[1], node.Operands.Exprs[2:]...)
	*n = *(eq.(*exprEq))
	return nil
}

type exprIn struct {
	Kind     string `json:"kind"`
	Operands []Expr `json:"operands"`
}

// In creates an expression that takes at least two operands and evaluates to true if the first one is equal to one of the others.
func In(v, e1 Expr, eN ...Expr) Expr {
	return &exprIn{
		Kind:     "in",
		Operands: append([]Expr{v, e1}, eN...),
	}
}

func (n *exprIn) Eval(params ParamGetter) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in In func")
	}

	toFind := n.Operands[0]
	vA, err := toFind.Eval(params)
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(n.Operands); i++ {
		vB, err := n.Operands[i].Eval(params)
		if err != nil {
			return nil, err
		}

		if vA.Equal(vB) {
			return BoolValue(true), nil
		}
	}

	return BoolValue(false), nil
}

func (n *exprIn) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Exprs) < 2 {
		return errors.New("invalid number of operands in In func")
	}

	in := In(node.Operands.Exprs[0], node.Operands.Exprs[1], node.Operands.Exprs[2:]...)

	*n = *(in.(*exprIn))

	return nil
}

type exprParam struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// StringParam creates an expression that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be a string. If not found it returns an error.
func StringParam(name string) Expr {
	return &exprParam{
		Kind: "param",
		Type: "string",
		Name: name,
	}
}

// BoolParam creates an expression that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be a boolean. If not found it returns an error.
func BoolParam(name string) Expr {
	return &exprParam{
		Kind: "param",
		Type: "bool",
		Name: name,
	}
}

// Int64Param creates an expression that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be an int64. If not found it returns an error.
func Int64Param(name string) Expr {
	return &exprParam{
		Kind: "param",
		Type: "int64",
		Name: name,
	}
}

// Float64Param creates an expression that looks up in the set of params passed during evaluation and returns the value
// of the variable that corresponds to the given name.
// The corresponding value must be a float64. If not found it returns an error.
func Float64Param(name string) Expr {
	return &exprParam{
		Kind: "param",
		Type: "float64",
		Name: name,
	}
}

func (n *exprParam) Eval(params ParamGetter) (*Value, error) {
	if params == nil {
		return nil, errors.New("params is nil")
	}

	switch n.Type {
	case "string":
		v, err := params.GetString(n.Name)
		if err != nil {
			return nil, err
		}
		return StringValue(v), nil
	case "bool":
		v, err := params.GetBool(n.Name)
		if err != nil {
			return nil, err
		}
		return BoolValue(v), nil
	case "int64":
		v, err := params.GetInt64(n.Name)
		if err != nil {
			return nil, err
		}
		return Int64Value(v), nil
	case "float64":
		v, err := params.GetFloat64(n.Name)
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
func (v *Value) Eval(ParamGetter) (*Value, error) {
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
