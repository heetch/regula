package store

import (
	"context"

	"github.com/heetch/rules-engine/rule"
)

// Store manages the storage of rulesets.
type Store interface {
	// All returns all the rulesets entries from the store.
	All(context.Context) ([]RulesetEntry, error)
}

// RulesetEntry holds a ruleset and its metadata.
type RulesetEntry struct {
	Name    string
	Ruleset *rule.Ruleset
}
