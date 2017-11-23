package mock

import (
	"github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
)

// Store ...
type Store struct {
	namespace string
	ruleSets  map[string]rule.Ruleset
}

// NewStore ...
func NewStore(namespace string, ruleSets map[string]rule.Ruleset) *Store {
	return &Store{
		namespace: namespace,
		ruleSets:  ruleSets,
	}
}

// Get ...
func (s *Store) Get(key string) (rule.Ruleset, error) {
	rs, ok := s.ruleSets[key]
	if !ok {
		err := rules.ErrRulesetNotFound
		return nil, err
	}

	return rs, nil
}
