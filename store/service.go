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

// RulesetService manages rulesets.
type RulesetService interface {
	// Get returns the ruleset related to the given path. By default, it returns the latest one.
	// It returns the related ruleset version if it's specified.
	Get(ctx context.Context, path, version string) (*RulesetEntry, error)
	// List returns the latest version of each ruleset under the given prefix.
	// If the prefix is empty, it returns entries from the beginning following the lexical order.
	// The listing can be customised using the ListOptions type.
	List(ctx context.Context, prefix string, opt *ListOptions) (*RulesetEntries, error)
	// Watch a prefix for changes and return a list of events.
	Watch(ctx context.Context, prefix string, revision string) (*RulesetEvents, error)
	// Put is used to store a ruleset version.
	Put(ctx context.Context, path string, ruleset *regula.Ruleset) (*RulesetEntry, error)
	// Eval evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
	Eval(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error)
	// EvalVersion evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
	EvalVersion(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
}

// ListOptions contains list options.
// If the Limit is lower or equal to 0 or greater than 100, it will be set to 50 by default.
type ListOptions struct {
	Limit         int
	ContinueToken string
	PathsOnly     bool // return only the paths of the rulesets
	AllVersions   bool // return all versions of each rulesets
}

// RulesetEntry holds a ruleset and its metadata.
type RulesetEntry struct {
	Path      string
	Version   string
	Ruleset   *regula.Ruleset
	Signature *regula.Signature
	Versions  []string
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
