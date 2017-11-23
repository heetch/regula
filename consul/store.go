package consul

import (
	"encoding/json"
	"strings"

	"github.com/heetch/rules-engine"

	"github.com/hashicorp/consul/api"
	"github.com/heetch/rules-engine/rule"
)

// Store holds a collection of rules by their keys
type Store struct {
	ruleSets map[string]rule.Ruleset
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

	m := map[string]rule.Ruleset{}

	for _, v := range pairs {
		rs, err := parseRuleset(v.Value)
		if err != nil {
			return nil, err
		}

		k := strings.TrimLeft(v.Key, keyPrefix)
		m["/"+k] = rs
	}

	return &Store{ruleSets: m}, nil
}

// Get returns a rule-set based on a given key
func (s *Store) Get(key string) (rule.Ruleset, error) {
	rs, ok := s.ruleSets["/"+key]

	if !ok {
		return nil, rules.ErrRulesetNotFound
	}

	return rs, nil
}

func parseRuleset(b []byte) (rule.Ruleset, error) {
	var rs rule.Ruleset
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, err
	}

	return rs, nil
}
