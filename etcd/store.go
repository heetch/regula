package etcd

import (
	"context"
	"encoding/json"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
	"github.com/pkg/errors"
)

const (
	timeout = 5 * time.Second
)

type Store struct {
	rulesets map[string]*rule.Ruleset
}

func NewStore(client *clientv3.Client, keyPrefix string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := client.Get(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve keys under the given keyPrefix")
	}

	m := make(map[string]*rule.Ruleset)

	for _, kv := range resp.Kvs {
		var rs rule.Ruleset
		if err := json.Unmarshal(kv.Value, &rs); err != nil {
			return nil, err
		}

		k := strings.TrimLeft(string(kv.Key), keyPrefix)
		m[path.Join("/", k)] = &rs
	}

	return &Store{rulesets: m}, nil
}

// Get returns a ruleset based on a given key.
func (s *Store) Get(key string) (*rule.Ruleset, error) {
	rs, ok := s.rulesets[path.Join("/", key)]
	if !ok {
		return nil, rules.ErrRulesetNotFound
	}

	return rs, nil
}
