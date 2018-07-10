package etcd

import (
	"context"
	"encoding/json"
	ppath "path"

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
	resp, err := s.Client.KV.Get(ctx, ppath.Join(s.Namespace, prefix), clientv3.WithPrefix())
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

// One returns the ruleset entry which corresponds to the given path.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *Store) One(ctx context.Context, path string) (*store.RulesetEntry, error) {
	resp, err := s.Client.KV.Get(ctx, ppath.Join(s.Namespace, path))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch the entry: %s", path)
	}

	// Count will be 0 if the path doesn't exist or if it's not a ruleset.
	if resp.Count == 0 {
		return nil, store.ErrNotFound
	}

	var entry store.RulesetEntry
	err = json.Unmarshal(resp.Kvs[0].Value, &entry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal entry")
	}

	return &entry, nil
}
