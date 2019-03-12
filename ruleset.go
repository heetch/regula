package regula

import (
	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

// A Ruleset is a list of rules.
type Ruleset struct {
	Path      string
	Version   string
	Rules     []*rule.Rule
	Signature *Signature
	Versions  []string
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
			tp, ok := signature.Params[p.Name]
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
	Params     map[string]string `json:"params"` // TODO(asdine) rename to Params
}

// Validate return type and parameters types.
func (s *Signature) Validate() error {
	switch s.ReturnType {
	case "string", "bool", "int64", "float64":
	default:
		return errors.New("unsupported return type")
	}

	for name, tp := range s.Params {
		switch tp {
		case "string", "bool", "int64", "float64":
		default:
			return errors.Errorf("unsupported param type '%s' for param '%s'", tp, name)
		}
	}

	return nil
}
