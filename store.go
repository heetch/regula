package rules

import "github.com/heetch/rules-engine/rule"

// Store ...
type Store interface {
	Get(key string) (*rule.Ruleset, error)
	FetchAll() error
}
