package consul

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/hashicorp/consul/api"
	rules "github.com/heetch/rules-engine"
)

// Store holds a collection of rules by their keys
type Store struct {
	ruleSets map[string]rules.RuleSet
}

// NewStore returns a Consul backed store, scoped on a given `keyPrefix`
func NewStore(consulAddr string, keyPrefix string) (*Store, error) {
	conf := &api.Config{
		Scheme:  "http",
		Address: consulAddr,
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	pairs, _, err := client.KV().List(keyPrefix, nil)
	if err != nil {
		return nil, err
	}

	m := map[string]rules.RuleSet{}

	for _, v := range pairs {
		rs, err := parseRuleSet(v.Value)
		if err != nil {
			return nil, err
		}

		k := strings.TrimLeft(v.Key, keyPrefix)
		m["/"+k] = rs
	}

	return &Store{ruleSets: m}, nil
}

// Get returns a rule-set based on a given key
func (s *Store) Get(key string) (rules.RuleSet, error) {
	rs, ok := s.ruleSets["/"+key]

	if !ok {
		return nil, errors.New("Key not found")
	}

	return rs, nil
}

func parseRuleSet(b []byte) (rules.RuleSet, error) {
	var rs rules.RuleSet
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, err
	}

	return rs, nil
}
