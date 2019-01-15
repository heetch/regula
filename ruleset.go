package regula

import (
	"encoding/json"
	"errors"

	"github.com/heetch/regula/param"
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
	rs := Ruleset{
		Rules: rules,
		Type:  typ,
	}

	err := rs.validate()
	if err != nil {
		return nil, err
	}

	return &rs, nil
}

// Eval evaluates every rule of the ruleset until one matches.
// It returns rule.ErrNoMatch if no rule matches the given context.
func (r *Ruleset) Eval(params param.Params) (*rule.Value, error) {
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

	return r.validate()
}

// Params returns a list of all the parameters used in all the underlying rules.
func (r *Ruleset) Params() []rule.Param {
	bm := make(map[string]bool)
	var params []rule.Param

	for _, rl := range r.Rules {
		ps := rl.Params()
		for _, p := range ps {
			if !bm[p.Name] {
				params = append(params, p)
				bm[p.Name] = true
			}
		}
	}

	return params
}

func (r *Ruleset) validate() error {
	paramTypes := make(map[string]string)

	for _, rl := range r.Rules {
		if rl.Result.Type != r.Type {
			return ErrRulesetIncoherentType
		}

		ps := rl.Params()
		for _, p := range ps {
			tp, ok := paramTypes[p.Name]
			if ok {
				if p.Type != tp {
					return ErrRulesetIncoherentType
				}
			} else {
				paramTypes[p.Name] = p.Type
			}
		}
	}

	return nil
}

// Signature represents the signature of a ruleset.
type Signature struct {
	ReturnType string
	ParamTypes map[string]string
}

// NewSignature returns the Signature of the given ruleset.
func NewSignature(rs *Ruleset) *Signature {
	pt := make(map[string]string)
	for _, p := range rs.Params() {
		pt[p.Name] = p.Type
	}

	return &Signature{
		ParamTypes: pt,
		ReturnType: rs.Type,
	}
}
