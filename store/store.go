package store

import (
	"context"
	"errors"

	"github.com/heetch/regula/rule"
)

// Errors
var (
	ErrNotFound = errors.New("not found")
)

// Store manages the storage of rulesets.
type Store interface {
	// List returns all the rulesets entries under the given prefix.
	List(ctx context.Context, prefix string) ([]RulesetEntry, error)
	// Latest returns the latest version of the ruleset entry which corresponds to the given path.
	Latest(ctx context.Context, path string) (*RulesetEntry, error)
	// OneByVersion returns the ruleset entry which corresponds to the given path at the given version.
	OneByVersion(ctx context.Context, path, version string) (*RulesetEntry, error)
	// Watch a prefix for changes and return a list of events.
	Watch(ctx context.Context, prefix string) ([]Event, error)
	// Put is used to store a ruleset version.
	Put(ctx context.Context, path string, ruleset *rule.Ruleset) (*RulesetEntry, error)
}

// RulesetEntry holds a ruleset and its metadata.
type RulesetEntry struct {
	Path    string
	Version string
	Ruleset *rule.Ruleset
}

// List of possible events executed against a ruleset.
const (
	PutEvent    = "PUT"
	DeleteEvent = "DELETE"
)

// Event describes an event that occured on a ruleset.
type Event struct {
	Type    string
	Path    string
	Ruleset *rule.Ruleset
}
