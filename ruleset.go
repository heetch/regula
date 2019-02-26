package regula

import (
	"encoding/json"

	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

// A Ruleset is list of rules.
type Ruleset struct {
	Rules []*rule.Rule `json:"rules"`
}

// NewRuleset creates a ruleset.
func NewRuleset(rules ...*rule.Rule) *Ruleset {
	rs := Ruleset{
		Rules: rules,
	}

	return &rs
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

	return nil
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

// ValidateSignature validates the ruleset against the given signature.
func (r *Ruleset) ValidateSignature(signature *Signature) error {
	if err := signature.Validate(); err != nil {
		return err
	}

	for _, rl := range r.Rules {
		if rl.Result.Type != signature.ReturnType {
			return rerrors.ErrSignatureMismatch
		}

		ps := rl.Params()
		for _, p := range ps {
			tp, ok := signature.ParamTypes[p.Name]
			if !ok || p.Type != tp {
				return rerrors.ErrSignatureMismatch
			}
		}
	}

	return nil
}

// Signature represents the signature of a ruleset.
type Signature struct {
	ReturnType string            `json:"returnType"`
	ParamTypes map[string]string `json:"paramTypes"` // TODO(asdine) rename to Params
}

// NewSignature create
func NewSignature() *Signature {
	return &Signature{
		ParamTypes: make(map[string]string),
	}
}

// StringP adds a string param to the signature.
func (s *Signature) StringP(name string) *Signature {
	s.ParamTypes[name] = "string"
	return s
}

// Int64P adds an int64 param to the signature.
func (s *Signature) Int64P(name string) *Signature {
	s.ParamTypes[name] = "int64"
	return s
}

// Float64P adds a float64 param to the signature.
func (s *Signature) Float64P(name string) *Signature {
	s.ParamTypes[name] = "float64"
	return s
}

// BoolP adds a bool param to the signature.
func (s *Signature) BoolP(name string) *Signature {
	s.ParamTypes[name] = "bool"
	return s
}

// ReturnsString sets the return type to string.
func (s *Signature) ReturnsString() *Signature {
	s.ReturnType = "string"
	return s
}

// ReturnsInt64 sets the return type to int64.
func (s *Signature) ReturnsInt64() *Signature {
	s.ReturnType = "int64"
	return s
}

// ReturnsFloat64 sets the return type to float64.
func (s *Signature) ReturnsFloat64() *Signature {
	s.ReturnType = "float64"
	return s
}

// ReturnsBool sets the return type to bool.
func (s *Signature) ReturnsBool() *Signature {
	s.ReturnType = "bool"
	return s
}

// Validate return type and parameters types.
func (s *Signature) Validate() error {
	switch s.ReturnType {
	case "string", "bool", "int64", "float64":
	default:
		return errors.New("unsupported return type")
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
