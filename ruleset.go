package regula

import (
	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

// A Ruleset is a list of rules and their metadata.
type Ruleset struct {
	Path      string       `json:"path"`
	Version   string       `json:"version,omitempty"`
	Rules     []*rule.Rule `json:"rules,omitempty"`
	Signature *Signature   `json:"signature,omitempty"`
	Versions  []string     `json:"versions,omitempty"`
}

// NewRuleset creates a ruleset.
func NewRuleset(rules ...*rule.Rule) *Ruleset {
	return &Ruleset{
		Rules: rules,
	}
}

// Eval evaluates the ruleset, evaluating every rule until one matches.
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
