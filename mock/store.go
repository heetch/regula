package mock

import (
	"errors"

	rules "github.com/heetch/rules-engine"
)

// Store ...
type Store struct {
	namespace string
	ruleSets  map[string]rules.RuleSet
}

// NewStore ...
func NewStore(namespace string, ruleSets map[string]rules.RuleSet) *Store {
	return &Store{
		namespace: namespace,
		ruleSets:  ruleSets,
	}
}

// Get ...
func (s *Store) Get(key string) (rules.RuleSet, error) {
	rs, ok := s.ruleSets[key]
	if !ok {
		err := errors.New("Key not found")
		return nil, err
	}

	return rs, nil
}

// FetchAll ...
func (s *Store) FetchAll() error {
	return nil
}
