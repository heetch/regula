package regula

import (
	"encoding/json"
	"errors"

	"github.com/heetch/regula/rule"
)

// A Ruleset is list of rules that must return the same type.
type Ruleset struct {
	Rules []*rule.Rule `json:"rules"`
	Type  string       `json:"type"`
}

// NewStringRuleset creates a ruleset which rules all return a string otherwise
// ErrRulesetIncoherentType is returned.
func NewStringRuleset(rules ...*rule.Rule) (*Ruleset, error) {
	return newRuleset("string", rules...)
}

// NewBoolRuleset creates a ruleset which rules all return a bool otherwise
// ErrRulesetIncoherentType is returned.
func NewBoolRuleset(rules ...*rule.Rule) (*Ruleset, error) {
	return newRuleset("bool", rules...)
}

// NewInt64Ruleset creates a ruleset which rules all return an int64 otherwise
// ErrRulesetIncoherentType is returned.
func NewInt64Ruleset(rules ...*rule.Rule) (*Ruleset, error) {
	return newRuleset("int64", rules...)
}

// NewFloat64Ruleset creates a ruleset which rules all return an float64 otherwise
// ErrRulesetIncoherentType is returned.
func NewFloat64Ruleset(rules ...*rule.Rule) (*Ruleset, error) {
	return newRuleset("float64", rules...)
}

func newRuleset(typ string, rules ...*rule.Rule) (*Ruleset, error) {
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
func (r *Ruleset) Eval(params rule.Params) (*rule.Value, error) {
	for _, rl := range r.Rules {
		res, err := rl.Eval(params)
		if err != rule.ErrNoMatch {
			return res, err
		}
	}

	return nil, rule.ErrNoMatch
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Ruleset) UnmarshalJSON(data []byte) error {
	type ruleset Ruleset
	if err := json.Unmarshal(data, (*ruleset)(r)); err != nil {
		return err
	}

	if r.Type != "string" && r.Type != "bool" && r.Type != "int64" && r.Type != "float64" {
		return errors.New("unsupported ruleset type")
	}

	_, err := newRuleset(r.Type, r.Rules...)
	return err
}
