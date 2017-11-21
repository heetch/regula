package rules

import (
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
)

// Rule represents the AST of a single rule.
type Rule struct {
	Root   Node
	Result Result
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Rule) UnmarshalJSON(data []byte) error {
	tree := struct {
		Root   json.RawMessage `json:"root"`
		Result Result          `json:"result"`
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
	Value string
	Type  string
}

type Node interface {
}

type NodeEq struct {
	Kind     string
	Operands []Node
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NodeEq) UnmarshalJSON(data []byte) error {
	node := struct {
		Kind     string   `json:"kind"`
		Operands operands `json:"operands"`
	}{}

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	n.Kind = node.Kind
	n.Operands = node.Operands.Nodes

	return nil
}

func parseNode(kind string, data []byte) (Node, error) {
	var n Node
	var err error

	switch kind {
	case "eq":
		var eq NodeEq
		n = &eq
		err = eq.UnmarshalJSON(data)
	default:
		err = errors.New("unknown kind")
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
