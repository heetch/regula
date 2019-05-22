// Package api provides types and interfaces that define how the Regula API is working.
package api

import (
	"context"

	"github.com/heetch/regula"
	"github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
)

// API errors.
const (
	ErrRulesetNotFound    = errors.Error("ruleset not found")
	ErrRulesetNotModified = errors.Error("not modified")
	ErrSignatureNotFound  = errors.Error("signature not found")
	ErrInvalidCursor      = errors.Error("invalid cursor")
	ErrAlreadyExists      = errors.Error("already exists")
)

// RulesetService is a service managing rulesets.
type RulesetService interface {
	// Create a ruleset using a signature.
	Create(ctx context.Context, path string, signature *regula.Signature) error
	// Put is used to add a new version of the rules to a ruleset.
	Put(ctx context.Context, path string, rules []*rule.Rule) (string, error)
	// Get returns a ruleset alongside its metadata.
	Get(ctx context.Context, path, version string) (*regula.Ruleset, error)
	// List returns the list of all rulesets paths.
	// The listing is paginated and can be customised using the ListOptions type.
	List(ctx context.Context, opt ListOptions) (*Rulesets, error)
	// Watch a prefix for changes and return a list of events.
	Watch(ctx context.Context, prefix string, revision string) (*RulesetEvents, error)
	// Eval evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
	Eval(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
}

// ListOptions is used to customize the List output.
type ListOptions struct {
	Limit  int    // If the Limit is lower or equal to 0 or greater than 100, it will be set to 50 by default.
	Cursor string // Pagination cursor. If empty the list starts from the beginning.
}

// GetLimit returns a limit that is between 1 and 100.
// If Limit is lower of equal to zero, it returns 50.
// If Limit is bigger than 100, it returns 100.
func (l *ListOptions) GetLimit() int {
	if l.Limit <= 0 {
		return 50
	}
	if l.Limit > 100 {
		return 100
	}

	return l.Limit
}

// Rulesets holds a list of rulesets.
type Rulesets struct {
	Paths    []string `json:"paths"`
	Revision string   `json:"revision"`         // revision when the request was applied
	Cursor   string   `json:"cursor,omitempty"` // cursor of the next page, if any
}

// List of possible events executed against a ruleset.
const (
	RulesetPutEvent = "PUT"
)

// RulesetEvent describes an event that occured on a ruleset.
type RulesetEvent struct {
	Type    string       `json:"type"`
	Path    string       `json:"path"`
	Version string       `json:"version"`
	Rules   []*rule.Rule `json:"rules"`
}

// RulesetEvents holds a list of events occured on a group of rulesets.
type RulesetEvents struct {
	Events   []RulesetEvent
	Revision string
	Timeout  bool // indicates if the watch did timeout
}
