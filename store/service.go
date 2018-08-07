package store

import (
	"context"
	"errors"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
)

// Errors.
var (
	ErrNotFound    = errors.New("not found")
	ErrNotModified = errors.New("not modified")
	// ErrBadRulesetName is returned if the name of the ruleset is badly formatted.
	ErrBadRulesetName = errors.New("bad ruleset name")
	// ErrBadParameterName is returned if the name of the parameter is badly formatted.
	ErrBadParameterName = errors.New("bad parameter name")
)

// RulesetService manages rulesets.
type RulesetService interface {
	// List returns all the rulesets entries under the given prefix.
	List(ctx context.Context, prefix string) (*RulesetEntries, error)
	// Latest returns the latest version of the ruleset entry which corresponds to the given path.
	Latest(ctx context.Context, path string) (*RulesetEntry, error)
	// OneByVersion returns the ruleset entry which corresponds to the given path at the given version.
	OneByVersion(ctx context.Context, path, version string) (*RulesetEntry, error)
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
