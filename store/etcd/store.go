package etcd

import (
	"context"
	"encoding/json"
	"log"

	"github.com/heetch/rules-engine/store"
	"github.com/pkg/errors"

	"github.com/coreos/etcd/clientv3"
)

var _ store.Store = new(Store)

// Store manages the storage of rulesets in etcd.
type Store struct {
	Logger    *log.Logger
	Client    *clientv3.Client
	Namespace string
}

// All returns all the rulesets entries from the store.
func (s *Store) All(ctx context.Context) ([]store.RulesetEntry, error) {
	resp, err := s.Client.KV.Get(ctx, s.Namespace, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all entries")
	}

	entries := make([]store.RulesetEntry, len(resp.Kvs))
	for i, pair := range resp.Kvs {
		err = json.Unmarshal(pair.Value, &entries[i])
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}
	}

	return entries, nil
}
