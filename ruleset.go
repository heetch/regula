package regula

import (
	"encoding/json"
	"errors"

	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
)

// A Ruleset is list of rules that must return the same type.
type Ruleset struct {
	Signature *Signature   `json:"signature"`
	Rules     []*rule.Rule `json:"rules"`
}

// NewRuleset creates a ruleset after having made sure that all the rules
// satisfy the given signature.
func NewRuleset(sig *Signature, rules ...*rule.Rule) (*Ruleset, error) {
	rs := Ruleset{
		Signature: sig,
		Rules:     rules,
	}

	return &rs, rs.validate()
}

// Eval evaluates every rule of the ruleset until one matches.
// It returns rule.ErrNoMatch if no rule matches the given context.
func (r *Ruleset) Eval(params rule.Params) (*rule.Value, error) {
	for _, rl := range r.Rules {
		res, err := rl.Eval(params)
		if err != rerrors.ErrNoMatch {
			return res, err
		}
	}

	return nil, rerrors.ErrNoMatch
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
			return rerrors.ErrRulesetIncoherentType
		}

		ps := rl.Params()
		for _, p := range ps {
			tp, ok := paramTypes[p.Name]
			if ok {
				if p.Type != tp {
					return rerrors.ErrRulesetIncoherentType
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
