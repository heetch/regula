package api

import (
	"context"
	"errors"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
)

// Common errors.
var (
	ErrRulesetNotFound      = errors.New("ruleset not found")
	ErrRulesetNotModified   = errors.New("not modified")
	ErrSignatureNotFound    = errors.New("signature not found")
	ErrInvalidContinueToken = errors.New("invalid continue token")
	ErrAlreadyExists        = errors.New("already exists")
)

// RulesetService manages rulesets.
type RulesetService interface {
	// Create a ruleset entry using a signature.
	Create(ctx context.Context, path string, signature *regula.Signature) error
	// Put is used to add rules to a ruleset. It creates a new version of the ruleset.
	Put(ctx context.Context, path string, rules []*rule.Rule) (*regula.Ruleset, error)
	// Get returns a ruleset alongside its metadata. By default, it returns the latest version.
	// If the version is not empty, the specified version is returned.
	Get(ctx context.Context, path, version string) (*regula.Ruleset, error)
	// List returns the latest version of each ruleset whose path starts by the given prefix.
	// If the prefix is empty, it returns all the entries following the lexical order.
	// The listing is paginated and can be customised using the ListOptions type.
	List(ctx context.Context, prefix string, opt *ListOptions) (*Rulesets, error)
	// Watch a prefix for changes and return a list of events.
	Watch(ctx context.Context, prefix string, revision string) (*RulesetEvents, error)
	// Eval evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
	Eval(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
}

// ListOptions contains list options.
// If the Limit is lower or equal to 0 or greater than 100, it will be set to 50 by default.
type ListOptions struct {
	Limit         int
	ContinueToken string
	PathsOnly     bool // return only the paths of the rulesets
	AllVersions   bool // return all versions of each rulesets
}

// Rulesets holds a list of rulesets.
type Rulesets struct {
	Rulesets []regula.Ruleset
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
	Rules   []*rule.Rule
}

// RulesetEvents holds a list of events occured on a group of rulesets.
type RulesetEvents struct {
	Events   []RulesetEvent
	Revision string
}
