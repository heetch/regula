package regula

import (
	"encoding/json"

	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
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
	if r.Signature == nil {
		return errors.New("missing signature")
	}

	if err := r.Signature.validate(); err != nil {
		return err
	}

	for _, rl := range r.Rules {
		if rl.Result.Type != r.Signature.ReturnType {
			return rerrors.ErrSignatureMismatch
		}

		ps := rl.Params()
		for _, p := range ps {
			tp, ok := r.Signature.ParamTypes[p.Name]
			if !ok || p.Type != tp {
				return rerrors.ErrSignatureMismatch
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

// validate return type and parameters types.
func (s *Signature) validate() error {
	switch s.ReturnType {
	case "string", "bool", "int64", "float64":
	default:
		return errors.New("unsupported ruleset type")
	}

	for name, tp := range s.ParamTypes {
		switch tp {
		case "string", "bool", "int64", "float64":
		default:
			return errors.Errorf("unsupported param type '%s' for param '%s'", tp, name)
		}
	}

	return nil
}
