package rules

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
	Get(key string) (rule.Ruleset, error)
}
