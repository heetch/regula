package rule

import (
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
)

// A Ruleset is list of rules.
type Ruleset []Rule

// Rule represents the AST of a single rule.
type Rule struct {
	Root   Operator `json:"root"`
	Result *Result  `json:"result"`
}

// New rule.
func New(fn func() (Operator, error), res *Result) (*Rule, error) {
	op, err := fn()
	if err != nil {
		return nil, err
	}

	return &Rule{
		Root:   op,
		Result: res,
	}, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Rule) UnmarshalJSON(data []byte) error {
	tree := struct {
		Root   json.RawMessage `json:"root"`
		Result *Result         `json:"result"`
	}{}

	err := json.Unmarshal(data, &tree)
	if err != nil {
		return err
	}

	res := gjson.Get(string(tree.Root), "kind")
	n, err := parseOperator(res.Str, []byte(tree.Root))
	if err != nil {
		return err
	}

	r.Root = n
	r.Result = tree.Result
	return err
}

// Result contains the value to return if the rule is matched.
type Result struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

// Returns specifies the result returned by the rule if matched.
func Returns(value, typ string) *Result {
	return &Result{
		Value: value,
		Type:  typ,
	}
}

// Operator represents a rule operator.
type Operator interface {
}

// Operand is used by an operator.
type Operand interface {
}

func parseOperand(kind string, data []byte) (Operand, error) {
	var n Operand
	var err error

	switch kind {
	case "value":
		var v OpValue
		n = &v
		err = json.Unmarshal(data, &v)
	case "variable":
		var v OpVariable
		n = &v
		err = json.Unmarshal(data, &v)
	case "true":
		var v OpTrue
		n = &v
		err = json.Unmarshal(data, &v)
	default:
		err = errors.New("unknown operand kind")
	}

	return n, err
}

func parseOperator(kind string, data []byte) (Operator, error) {
	var n Operator
	var err error

	switch kind {
	case "eq":
		var eq OpEq
		n = &eq
		err = eq.UnmarshalJSON(data)
	case "in":
		var in OpIn
		n = &in
		err = in.UnmarshalJSON(data)
	default:
		err = errors.New("unknown operator kind")
	}

	return n, err
}

type operands struct {
	Ops   []json.RawMessage `json:"operands"`
	Nodes []Operand
}

func (o *operands) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &o.Ops)
	if err != nil {
		return err
	}

	for _, op := range o.Ops {
		r := gjson.Get(string(op), "kind")
		n, err := parseOperand(r.Str, []byte(op))
		if err != nil {
			return err
		}

		o.Nodes = append(o.Nodes, n)
	}

	return nil
}

type nodeOps struct {
	Kind     string   `json:"kind"`
	Operands operands `json:"operands"`
}

// OpEq represents the Eq operator.
type OpEq struct {
	Kind     string    `json:"kind"`
	Operands []Operand `json:"operands"`
}

// Eq creates an Eq operator.
func Eq(ops ...Operand) func() (Operator, error) {
	return func() (Operator, error) {
		return &OpEq{
			Kind:     "eq",
			Operands: ops,
		}, nil
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *OpEq) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	o.Kind = node.Kind
	o.Operands = node.Operands.Nodes

	return nil
}

// OpIn represents the In operator.
type OpIn struct {
	Kind     string    `json:"kind"`
	Operands []Operand `json:"operands"`
}

// In creates an In operator.
func In(v *OpVariable, vals ...*OpValue) func() (Operator, error) {
	ops := make([]Operand, len(vals)+1)
	ops[0] = v
	for i := range vals {
		ops[i+1] = vals[i]
	}

	return func() (Operator, error) {
		o := &OpIn{
			Kind:     "in",
			Operands: ops,
		}
		return o, nil
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *OpIn) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	n.Kind = node.Kind
	n.Operands = node.Operands.Nodes

	return nil
}

// Variable creates the variable operand.
func Variable(name, typ string) *OpVariable {
	return &OpVariable{
		Kind: "variable",
		Type: typ,
		Name: name,
	}
}

// OpVariable represents the variable operand.
type OpVariable struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// Value creates the value operand.
func Value(value, typ string) *OpValue {
	return &OpValue{
		Kind:  "value",
		Type:  typ,
		Value: value,
	}
}

// OpValue represents the value operand.
type OpValue struct {
	Kind  string `json:"kind"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// True creates the true operand.
func True() *OpTrue {
	return &OpTrue{
		Kind: "true",
	}
}

// OpTrue represents the true operand.
type OpTrue struct {
	Kind string `json:"kind"`
}
