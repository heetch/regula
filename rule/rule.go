// Package rule provides primitives for creating and evaluating logical rules.
package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

var (
	// ErrRulesetIncoherentType is returned when a ruleset contains rules of different types
	ErrRulesetIncoherentType = errors.New("types in ruleset are incoherent")
)

// Params is passed on rule evaluation.
type Params map[string]string

// Rule represents the AST of a single rule.
type Rule struct {
	Root   Node    `json:"root"`
	Result *Result `json:"result"`
}

// New rule.
func New(node Node, res *Result) *Rule {
	return &Rule{
		Root:   node,
		Result: res,
	}
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

	if tree.Result.Type == "" {
		return errors.New("invalid rule result type")
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

// Eval evaluates the rule against the given context.
// If it matches it returns a result, otherwise it returns ErrNoMatch
// or any encountered error.
func (r *Rule) Eval(params Params) (*Result, error) {
	value, err := r.Root.Eval(params)
	if err != nil {
		return nil, err
	}

	if value.Type != "bool" {
		return nil, errors.New("invalid rule returning non boolean value")
	}

	ok, err := strconv.ParseBool(value.Data)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrNoMatch
	}

	return r.Result, nil
}

// Result contains the value to return if the rule is matched.
type Result struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

// ReturnsStr specifies the string result to be returned by the rule if matched.
func ReturnsStr(value string) *Result {
	return &Result{
		Value: value,
		Type:  "string",
	}
}

// ReturnsBool specifies the bool result to be returned by the rule if matched.
func ReturnsBool(value bool) *Result {
	return &Result{
		Value: fmt.Sprintf("%v", value),
		Type:  "bool",
	}
}

// A Ruleset is list of rules and a type
type Ruleset struct {
	Rules []*Rule `json:"rules"`
	Type  string  `json:"type"`
}

// NewRuleset creates a ruleset based on given type and rules.
// All rules must have the same return type as specified, otherwise a
// ErrRulesetIncoherentType will be returned.
func NewRuleset(typ string, rules ...*Rule) (*Ruleset, error) {
	for _, r := range rules {
		if typ != r.Result.Type {
			return nil, ErrRulesetIncoherentType
		}
	}

	return &Ruleset{
		Rules: rules,
		Type:  typ,
	}, nil
}

// Eval evaluates every rule of the ruleset until one matches.
// It returns rule.ErrNoMatch if no rule matches the given context.
func (r *Ruleset) Eval(params Params) (*Result, error) {
	for _, rl := range r.Rules {
		res, err := rl.Eval(params)
		if err != ErrNoMatch {
			return res, err
		}
	}

	return nil, ErrNoMatch
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Ruleset) UnmarshalJSON(data []byte) error {
	type ruleset Ruleset
	if err := json.Unmarshal(data, (*ruleset)(r)); err != nil {
		return err
	}

	_, err := NewRuleset(r.Type, r.Rules...)
	return err
}
