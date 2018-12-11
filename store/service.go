package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
)

// Common errors.
var (
	ErrNotFound             = errors.New("not found")
	ErrNotModified          = errors.New("not modified")
	ErrInvalidContinueToken = errors.New("invalid continue token")
)

// ValidationError gives informations about the reason of failed validation.
type ValidationError struct {
	Field  string
	Value  string
	Reason string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s with value '%s': %s", v.Field, v.Value, v.Reason)
}

// IsValidationError indicates if the given error is a ValidationError pointer.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// RulesetService manages rulesets.
type RulesetService interface {
	// List returns the rulesets entries under the given prefix. if pathsOnly is set to true, only the rulesets paths are returned.
	// If the prefix is empty it returns entries from the beginning following the ascii ordering.
	// If the given limit is lower or equal to 0 or greater than 100, it returns 50 entries.
	List(ctx context.Context, prefix string, limit int, continueToken string, pathsOnly bool) (*RulesetEntries, error)
	// Watch a prefix for changes and return a list of events.
	Watch(ctx context.Context, prefix string, revision string) (*RulesetEvents, error)
	// Put is used to store a ruleset version.
	Put(ctx context.Context, path string, ruleset *regula.Ruleset) (*RulesetEntry, error)
	// Eval evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
	Eval(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error)
	// EvalVersion evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
	EvalVersion(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
}

// RulesetEntry holds a ruleset and its metadata.
type RulesetEntry struct {
	Path      string
	Version   string
	Ruleset   *regula.Ruleset
	Signature *Signature
}

// RulesetEntries holds a list of ruleset entries.
type RulesetEntries struct {
	Entries  []RulesetEntry
	Revision string // revision when the request was applied
	Continue string // token of the next page, if any
}

// List of possible events executed against a ruleset.
const (
	RulesetPutEvent = "PUT"
)

// RulesetEvent describes an event that occured on a ruleset.
type RulesetEvent struct {
	Type    string
	Path    string
	Version string
	Ruleset *regula.Ruleset
}

// RulesetEvents holds a list of events occured on a group of rulesets.
type RulesetEvents struct {
	Events   []RulesetEvent
	Revision string
}

// Signature represents the signature of a ruleset.
type Signature struct {
	ReturnType string
	ParamTypes map[string]string
}

// NewSignature returns the Signature of the given ruleset.
func NewSignature(rs *regula.Ruleset) *Signature {
	pt := make(map[string]string)
	for _, p := range rs.Params() {
		pt[p.Name] = p.Type
	}

	return &Signature{
		ParamTypes: pt,
		ReturnType: rs.Type,
	}
}

// MatchWith checks if the given signature match the actual one.
func (s *Signature) MatchWith(other *Signature) error {
	if s.ReturnType != other.ReturnType {
		return &ValidationError{
			Field:  "return type",
			Value:  other.ReturnType,
			Reason: fmt.Sprintf("signature mismatch: return type must be of type %s", s.ReturnType),
		}
	}

	for name, tp := range other.ParamTypes {
		stp, ok := s.ParamTypes[name]
		if !ok {
			return &ValidationError{
				Field:  "param",
				Value:  name,
				Reason: "signature mismatch: unknown parameter",
			}
		}

		if tp != stp {
			return &ValidationError{
				Field:  "param type",
				Value:  tp,
				Reason: fmt.Sprintf("signature mismatch: param must be of type %s", stp),
			}
		}
	}

	return nil
}
