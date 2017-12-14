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

// Store is an etcd store that holds rulesets in memory.
type Store struct {
	rulesets map[string]*rule.Ruleset
}

// NewStore takes a connected etcd client, fetches all the rulesets under the given prefix and stores them in the returned store.
// Any leading slash found on keyPrefix is removed and a trailing slash is added automatically before usage.
func NewStore(client *clientv3.Client, keyPrefix string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	keyPrefix = path.Join(strings.TrimLeft(keyPrefix, "/"), "/")

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

// Get returns a memory stored ruleset based on a given key.
// No network round trip is perfomed during this call.
func (s *Store) Get(key string) (*rule.Ruleset, error) {
	rs, ok := s.rulesets[path.Join("/", key)]
	if !ok {
		return nil, rules.ErrRulesetNotFound
	}

	return rs, nil
}
