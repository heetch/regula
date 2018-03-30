package store

import (
	"github.com/heetch/rules-engine/rule"
	"github.com/pkg/errors"
)

var (
	// ErrRulesetNotFound must be returned when no ruleset is found in the store for a given key.
	ErrRulesetNotFound = errors.New("ruleset not found")
)

// Store manages the storage of rulesets.
type Store interface {
	// Get returns the ruleset associated with the given key.
	// If no ruleset is found for a given key, the implementation must return ErrRulesetNotFound.
	Get(key string) (*rule.Ruleset, error)
}
