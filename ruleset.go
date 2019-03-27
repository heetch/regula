package regula

import (
	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

// A Ruleset is a list of rules and their metadata.
type Ruleset struct {
	Path      string           `json:"path"`
	Signature *Signature       `json:"signature,omitempty"`
	Versions  []RulesetVersion `json:"versions,omitempty"`
}

// NewRuleset creates a ruleset.
func NewRuleset(rules ...*rule.Rule) *Ruleset {
	var rs Ruleset

	rs.Versions = append(rs.Versions, RulesetVersion{
		Version: "latest",
		Rules:   rules,
	})

	return &rs
}

// Eval evaluates the latest ruleset version, evaluating every rule until one matches.
// It returns rule.ErrNoMatch if no rule matches the given context.
func (r *Ruleset) Eval(params rule.Params) (*rule.Value, error) {
	if len(r.Versions) == 0 {
		return nil, rerrors.ErrNoMatch
	}

	for _, rl := range r.Versions[len(r.Versions)-1].Rules {
		res, err := rl.Eval(params)
		if err != rerrors.ErrNoMatch {
			return res, err
		}
	}

	return nil, rerrors.ErrRulesetVersionNotFound
}

// EvalVersion evaluates a version of the ruleset, evaluating every rule until one matches.
// It returns rule.ErrNoMatch if no rule matches the given context.
func (r *Ruleset) EvalVersion(version string, params rule.Params) (*rule.Value, error) {
	if len(r.Versions) == 0 {
		return nil, rerrors.ErrNoMatch
	}

	for _, rv := range r.Versions {
		if rv.Version == version {
			for _, rl := range rv.Rules {
				res, err := rl.Eval(params)
				if err != rerrors.ErrNoMatch {
					return res, err
				}
			}
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

// A RulesetVersion describes a version of a list of rules.
type RulesetVersion struct {
	Version string
	Rules   []*rule.Rule
}
