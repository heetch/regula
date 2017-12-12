package rule

import (
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
)

// A Node is a piece of the AST that denotes a construct occurring in the rule source code.
// Each node takes a set of params and evaluates to a value.
type Node interface {
	Eval(Params) (*Value, error)
}

func parseNode(kind string, data []byte) (Node, error) {
	var n Node
	var err error

	switch kind {
	case "eq":
		var eq nodeEq
		n = &eq
		err = eq.UnmarshalJSON(data)
	case "in":
		var in nodeIn
		n = &in
		err = in.UnmarshalJSON(data)
	case "not":
		var not nodeNot
		n = &not
		err = not.UnmarshalJSON(data)
	case "and":
		var and nodeAnd
		n = &and
		err = and.UnmarshalJSON(data)
	case "or":
		var or nodeOr
		n = &or
		err = or.UnmarshalJSON(data)
	case "value":
		var v nodeValue
		n = &v
		err = json.Unmarshal(data, &v)
	case "param":
		var v nodeParam
		n = &v
		err = json.Unmarshal(data, &v)
	case "true":
		var v nodeTrue
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

type nodeNot struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// Not creates a node that evaluates the given node n and returns its opposite.
// n must evaluate to a boolean.
func Not(n Node) Node {
	return &nodeNot{
		Kind:     "not",
		Operands: []Node{n},
	}
}

func (n *nodeNot) Eval(params Params) (*Value, error) {
	if len(n.Operands) < 1 {
		return nil, errors.New("invalid number of operands in not func")
	}

	op := n.Operands[0]
	v, err := op.Eval(params)
	if err != nil {
		return nil, err
	}

	if v.Type != "bool" {
		return nil, errors.New("invalid operand type for Not func")
	}

	if v.Equal(NewBoolValue(true)) {
		return NewBoolValue(false), nil
	}

	return NewBoolValue(true), nil
}

func (n *nodeNot) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Nodes) < 1 {
		return errors.New("invalid number of operands in not func")
	}

	n.Operands = node.Operands.Nodes
	n.Kind = "not"

	return nil
}

type nodeOr struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// Or creates a node that takes at least two nodes and evaluates to true if one of the nodes evaluates to true.
// All the given nodes must evaluate to a boolean.
func Or(v1, v2 Node, vN ...Node) Node {
	return &nodeOr{
		Kind:     "or",
		Operands: append([]Node{v1, v2}, vN...),
	}
}

func (n *nodeOr) Eval(params Params) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in or func")
	}

	opA := n.Operands[0]
	vA, err := opA.Eval(params)
	if err != nil {
		return nil, err
	}
	if vA.Type != "bool" {
		return nil, errors.New("invalid operand type for Or func")
	}

	if vA.Equal(NewBoolValue(true)) {
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

		if vB.Equal(NewBoolValue(true)) {
			return vB, nil
		}
	}

	return NewBoolValue(false), nil
}

func (n *nodeOr) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in or func")
	}

	or := Or(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)
	*n = *(or.(*nodeOr))
	return nil
}

type nodeAnd struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// And creates a node that takes at least two nodes and evaluates to true if all the nodes evaluate to true.
// All the given nodes must evaluate to a boolean.
func And(v1, v2 Node, vN ...Node) Node {
	return &nodeAnd{
		Kind:     "and",
		Operands: append([]Node{v1, v2}, vN...),
	}
}

func (n *nodeAnd) Eval(params Params) (*Value, error) {
	if len(n.Operands) < 2 {
		return nil, errors.New("invalid number of operands in or func")
	}

	opA := n.Operands[0]
	vA, err := opA.Eval(params)
	if err != nil {
		return nil, err
	}
	if vA.Type != "bool" {
		return nil, errors.New("invalid operand type for Or func")
	}

	if vA.Equal(NewBoolValue(false)) {
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

		if vB.Equal(NewBoolValue(false)) {
			return vB, nil
		}
	}

	return NewBoolValue(true), nil
}

func (n *nodeAnd) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in or func")
	}

	and := And(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)
	*n = *(and.(*nodeAnd))
	return nil
}

type nodeEq struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// Eq creates a node that takes at least two nodes and evaluates to true if all the nodes are equal.
func Eq(v1, v2 Node, vN ...Node) Node {
	return &nodeEq{
		Kind:     "eq",
		Operands: append([]Node{v1, v2}, vN...),
	}
}

func (n *nodeEq) Eval(params Params) (*Value, error) {
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

func (n *nodeEq) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in eq func")
	}

	eq := Eq(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)
	*n = *(eq.(*nodeEq))
	return nil
}

type nodeIn struct {
	Kind     string `json:"kind"`
	Operands []Node `json:"operands"`
}

// In creates a node that takes at least two nodes and evaluates to true if the first one is equal to one of the others.
func In(v, e1 Node, eN ...Node) Node {
	return &nodeIn{
		Kind:     "in",
		Operands: append([]Node{v, e1}, eN...),
	}
}

func (n *nodeIn) Eval(params Params) (*Value, error) {
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

func (n *nodeIn) UnmarshalJSON(data []byte) error {
	var node nodeOps

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	if len(node.Operands.Nodes) < 2 {
		return errors.New("invalid number of operands in in func")
	}

	in := In(node.Operands.Nodes[0], node.Operands.Nodes[1], node.Operands.Nodes[2:]...)

	*n = *(in.(*nodeIn))

	return nil
}

type nodeParam struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// ParamStr creates a node that looks up in the set of params passed during evaluation and returns the value of the variable that corresponds to the given name.
// The corresponding value must be a string. If not found it returns an error.
func ParamStr(name string) Node {
	return &nodeParam{
		Kind: "param",
		Type: "string",
		Name: name,
	}
}

func (n *nodeParam) Eval(params Params) (*Value, error) {
	val, ok := params[n.Name]
	if !ok {
		return nil, errors.New("param not found in given context")
	}

	return &Value{
		Type: n.Type,
		Data: val,
	}, nil
}

type nodeValue struct {
	Kind  string `json:"kind"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ValueStr creates a node that evaluates to a constant of type string.
func ValueStr(value string) Node {
	return &nodeValue{
		Kind:  "value",
		Type:  "string",
		Value: value,
	}
}

func (n *nodeValue) Eval(Params) (*Value, error) {
	return &Value{
		Type: n.Type,
		Data: n.Value,
	}, nil
}

type nodeTrue struct {
	Kind string `json:"kind"`
}

// True creates a node that always evaluates to true.
func True() Node {
	return &nodeTrue{
		Kind: "true",
	}
}

func (v *nodeTrue) Eval(Params) (*Value, error) {
	return NewBoolValue(true), nil
}
