package etcd

import (
	"context"
	"encoding/json"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/rules-engine/store"
	"github.com/pkg/errors"
)

var _ store.Store = new(Store)

// Store manages the storage of rulesets in etcd.
type Store struct {
	Client    *clientv3.Client
	Namespace string
}

// List returns all the rulesets entries under the given path.
func (s *Store) List(ctx context.Context, prefix string) ([]store.RulesetEntry, error) {
	resp, err := s.Client.KV.Get(ctx, path.Join(s.Namespace, prefix), clientv3.WithPrefix())
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
