package client

import (
	"context"

	"github.com/heetch/rules-engine/rule"
)

// A Client manages communication with the Rules Engine API.
type Client interface {
	// Get returns the ruleset associated with the given key.
	// If no ruleset is found for a given key, the implementation must return store.ErrRulesetNotFound.
	Get(ctx context.Context, key string) (*rule.Ruleset, error)
}
