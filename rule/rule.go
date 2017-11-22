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
	Root   Node    `json:"root"`
	Result *Result `json:"result"`
}

// New rule.
func New(fn func() (Node, error), res *Result) (*Rule, error) {
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
	n, err := parseNode(res.Str, []byte(tree.Root))
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

// ReturnsStr specifies the result returned by the rule if matched.
func ReturnsStr(value string) *Result {
	return &Result{
		Value: value,
		Type:  "string",
	}
}

// Node represents a rule Node.
type Node interface {
}

func parseNode(kind string, data []byte) (Node, error) {
	var n Node
	var err error

	switch kind {
	case "eq":
		var eq NodeEq
		n = &eq
		err = eq.UnmarshalJSON(data)
	case "in":
		var in NodeIn
		n = &in
		err = in.UnmarshalJSON(data)
	case "value":
		var v NodeValue
		n = &v
		err = json.Unmarshal(data, &v)
	case "variable":
		var v NodeVariable
		n = &v
		err = json.Unmarshal(data, &v)
	case "true":
		var v NodeTrue
		n = &v
		err = json.Unmarshal(data, &v)
	default:
		err = errors.New("unknown node kind")
	}

	return n, err
}

type operands struct {
	Ops   []json.RawMessage `json:"operands"`
	Nodes []Node
}

func (o *operands) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &o.Ops)
	if err != nil {
		return err
	}

	for _, op := range o.Ops {
		r := gjson.Get(string(op), "kind")
		n, err := parseNode(r.Str, []byte(op))
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

// NodeEq represents the Eq Node.
type NodeEq struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// Eq creates an Eq Node.
func Eq(v1, v2 Node, vN ...Node) func() (Node, error) {
	return func() (Node, error) {
		return &NodeEq{
			Kind:     "eq",
			Operands: append([]Node{v1, v2}, vN...),
		}, nil
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *NodeEq) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	o.Kind = node.Kind
	o.Operands = node.Operands.Nodes

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in eq func")
	}

	n, err := Eq(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)()
	if err != nil {
		return err
	}

	*o = *(n.(*NodeEq))

	return nil
}

// NodeIn represents the In Node.
type NodeIn struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// In creates an In Node.
func In(v, e1 Node, eN ...Node) func() (Node, error) {
	return func() (Node, error) {
		o := &NodeIn{
			Kind:     "in",
			Operands: append([]Node{v, e1}, eN...),
		}
		return o, nil
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (i *NodeIn) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	i.Kind = node.Kind

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in in func")
	}

	n, err := In(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)()
	if err != nil {
		return err
	}

	*i = *(n.(*NodeIn))

	return nil
}

// NodeVariable represents the variable node.
type NodeVariable struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// VarStr creates a variable node of type string.
func VarStr(name string) *NodeVariable {
	return &NodeVariable{
		Kind: "variable",
		Type: "string",
		Name: name,
	}
}

// NodeValue represents the value node.
type NodeValue struct {
	Kind  string `json:"kind"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ValStr creates a value node of type string.
func ValStr(value string) *NodeValue {
	return &NodeValue{
		Kind:  "value",
		Type:  "string",
		Value: value,
	}
}

// NodeTrue represents the true node.
type NodeTrue struct {
	Kind string `json:"kind"`
}

// True creates a true node.
func True() *NodeTrue {
	return &NodeTrue{
		Kind: "true",
	}
}
