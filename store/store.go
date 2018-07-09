package store

import (
	"context"
	"errors"

	"github.com/heetch/rules-engine/rule"
)

// Store errors
var (
	ErrNotFound = errors.New("not found")
)

// Store manages the storage of rulesets.
type Store interface {
	// List returns all the rulesets entries under the given prefix.
	List(ctx context.Context, prefix string) ([]RulesetEntry, error)

	// One returns the ruleset entry which corresponds to the given path.
	One(ctx context.Context, path string) (*RulesetEntry, error)
}

// RulesetEntry holds a ruleset and its metadata.
type RulesetEntry struct {
	Path    string        `json:"path"`
	Ruleset *rule.Ruleset `json:"ruleset"`
}
