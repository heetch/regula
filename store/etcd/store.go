package etcd

import (
	"context"
	"encoding/json"
	ppath "path"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/globalsign/mgo/bson"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
)

var _ store.Store = new(Store)

// Store manages the storage of rulesets in etcd.
type Store struct {
	Client    *clientv3.Client
	Namespace string
}

// List returns all the rulesets entries under the given prefix.
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
	// TODO fix how rulesets are get
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

// Put adds a version of the given ruleset using an uuid.
func (s *Store) Put(ctx context.Context, path string, ruleset *rule.Ruleset) (*store.RulesetEntry, error) {
	v := bson.NewObjectId().Hex()

	re := store.RulesetEntry{
		Path:    path,
		Version: v,
		Ruleset: ruleset,
	}

	raw, err := json.Marshal(&re)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode entry")
	}

	_, err = s.Client.KV.Put(ctx, ppath.Join(s.Namespace, path, v), string(raw))
	if err != nil {
		return nil, errors.Wrap(err, "failed to put entry")
	}

	return &re, nil
}

// Watch the given prefix for anything new.
func (s *Store) Watch(ctx context.Context, prefix string) ([]store.Event, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wc := s.Client.Watch(ctx, ppath.Join(s.Namespace, prefix), clientv3.WithPrefix())
	select {
	case wresp := <-wc:
		events := make([]store.Event, len(wresp.Events))
		for i, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				events[i].Type = store.PutEvent
			case mvccpb.DELETE:
				events[i].Type = store.DeleteEvent
			}

			events[i].Path = strings.TrimPrefix(s.Namespace, string(ev.Kv.Key))
			var e store.RulesetEntry
			err := json.Unmarshal(ev.Kv.Value, &e)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal entry")
			}
			events[i].Ruleset = e.Ruleset
		}

		return events, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
