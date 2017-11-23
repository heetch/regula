package rule

import (
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
)

// Node represents a rule Node.
type Node interface {
	Eval(Params) (*Value, error)
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
func Eq(v1, v2 Node, vN ...Node) *NodeEq {
	return &NodeEq{
		Kind:     "eq",
		Operands: append([]Node{v1, v2}, vN...),
	}
}

// Eval evaluates into true if all the operands are equal.
func (n *NodeEq) Eval(params Params) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in eq func")
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
			return NewBoolValue(false), nil
		}
	}

	return NewBoolValue(true), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NodeEq) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in eq func")
	}

	eq := Eq(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)
	*n = *eq
	return nil
}

// NodeIn represents the In Node.
type NodeIn struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// In creates an In Node.
func In(v, e1 Node, eN ...Node) *NodeIn {
	return &NodeIn{
		Kind:     "in",
		Operands: append([]Node{v, e1}, eN...),
	}
}

// Eval evaluates to true if the first operand is equal to one of the others.
func (n *NodeIn) Eval(params Params) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in eq func")
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
			return NewBoolValue(true), nil
		}
	}

	return NewBoolValue(false), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NodeIn) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in in func")
	}

	in := In(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)

	*n = *in

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

// Eval evaluates to the value of the variable contained in the given context.
// If not found it returns an error.
func (n *NodeVariable) Eval(params Params) (*Value, error) {
	val, ok := params[n.Name]
	if !ok {
		return nil, errors.New("variable not found in given context")
	}

	return &Value{
		Type: n.Type,
		Data: val,
	}, nil
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

// Eval evaluates into a value of the same type and value as the NodeValue.
func (n *NodeValue) Eval(Params) (*Value, error) {
	return &Value{
		Type: n.Type,
		Data: n.Value,
	}, nil
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

// Eval always evaluates to true.
func (v *NodeTrue) Eval(Params) (*Value, error) {
	return NewBoolValue(true), nil
}
