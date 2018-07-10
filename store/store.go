package store

import (
	"context"

	"github.com/heetch/rules-engine/rule"
)

// Store manages the storage of rulesets.
type Store interface {
	// List returns all the rulesets entries under the given prefix.
	List(ctx context.Context, prefix string) ([]RulesetEntry, error)
}

// RulesetEntry holds a ruleset and its metadata.
type RulesetEntry struct {
	Path    string        `json:"path"`
	Ruleset *rule.Ruleset `json:"ruleset"`
}
