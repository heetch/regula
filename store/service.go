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
	// List returns all the rulesets entries under the given prefix.
	// If the prefix is empty, it returns **all** the rulesets entries.
	// Instead, a limit option can be passed to return a subset of the rulesets.
	List(ctx context.Context, prefix string, limit int, continueToken string) (*RulesetEntries, error)
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
	Path    string
	Version string
	Ruleset *regula.Ruleset
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
