package etcd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"path"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/heetch/regula"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

// Store manages the storage of rulesets in etcd.
type Store struct {
	Client    *clientv3.Client
	Namespace string
}

// List returns all the rulesets entries under the given prefix.
func (s *Store) List(ctx context.Context, prefix string) (*store.RulesetEntries, error) {
	resp, err := s.Client.KV.Get(ctx, s.rulesetPath(prefix, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all entries")
	}

	var entries store.RulesetEntries
	entries.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	entries.Entries = make([]store.RulesetEntry, len(resp.Kvs))
	for i, pair := range resp.Kvs {
		err = json.Unmarshal(pair.Value, &entries.Entries[i])
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}
	}

	return &entries, nil
}

// Latest returns the latest version of the ruleset entry which corresponds to the given path.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *Store) Latest(ctx context.Context, path string) (*store.RulesetEntry, error) {
	if path == "" {
		return nil, store.ErrNotFound
	}

	resp, err := s.Client.KV.Get(ctx, s.rulesetPath(path, "")+"/", clientv3.WithLastKey()...)
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

// OneByVersion returns the ruleset entry which corresponds to the given path at the given version.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *Store) OneByVersion(ctx context.Context, path, version string) (*store.RulesetEntry, error) {
	if path == "" {
		return nil, store.ErrNotFound
	}

	resp, err := s.Client.KV.Get(ctx, s.rulesetPath(path, version))
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
func (s *Store) Put(ctx context.Context, path string, ruleset *regula.Ruleset) (*store.RulesetEntry, error) {
	// generate checksum from the ruleset for comparison purpose
	h := md5.New()
	err := json.NewEncoder(h).Encode(ruleset)
	if err != nil {
		return nil, err
	}
	checksum := string(h.Sum(nil))

	k, err := ksuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate version")
	}
	v := k.String()

	re := store.RulesetEntry{
		Path:    path,
		Version: v,
		Ruleset: ruleset,
	}

	raw, err := json.Marshal(&re)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode entry")
	}

	resp, err := s.Client.KV.Txn(ctx).
		If(
			// if last stored checksum equal the current one
			clientv3.Compare(clientv3.Value(s.checksumPath(path)), "=", checksum),
		).
		Then(
			// the latest ruleset is the same as this one
			// we fetch the latest ruleset entry
			clientv3.OpGet(s.rulesetPath(path, "")+"/", clientv3.WithLastKey()...),
		).
		Else(
			// if the checksum is different from the last one OR
			// if the checksum doesn't exist (first time we create a ruleset for this path)
			// we create a new version
			clientv3.OpPut(s.rulesetPath(path, v), string(raw)),
			// we create/update the checksum for the last ruleset version
			clientv3.OpPut(s.checksumPath(path), checksum),
		).
		Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to put entry")
	}

	// the checksum is the same as the last ruleset saved, we return an error
	// and the entry
	if resp.Succeeded {
		if len(resp.Responses) == 0 || resp.Responses[0].GetResponseRange().Count == 0 {
			return nil, errors.New("succeeded txn response not received")
		}

		var entry store.RulesetEntry
		err = json.Unmarshal(resp.Responses[0].GetResponseRange().Kvs[0].Value, &entry)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}

		return &entry, store.ErrNotModified
	}

	return &re, nil
}

// Watch the given prefix for anything new.
func (s *Store) Watch(ctx context.Context, prefix string, revision string) (*store.Events, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	if i, _ := strconv.ParseInt(revision, 10, 64); i > 0 {
		// watch from the next revision
		opts = append(opts, clientv3.WithRev(i+1))
	}

	wc := s.Client.Watch(ctx, s.rulesetPath(prefix, ""), opts...)
	for {
		select {
		case wresp := <-wc:
			if err := wresp.Err(); err != nil {
				return nil, errors.Wrapf(err, "failed to watch prefix: '%s'", prefix)
			}

			if len(wresp.Events) == 0 {
				continue
			}

			events := make([]store.Event, len(wresp.Events))
			for i, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					events[i].Type = store.PutEvent
				}

				var e store.RulesetEntry
				err := json.Unmarshal(ev.Kv.Value, &e)
				if err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal entry")
				}
				events[i].Path = e.Path
				events[i].Ruleset = e.Ruleset
				events[i].Version = e.Version
			}

			return &store.Events{
				Events:   events,
				Revision: strconv.FormatInt(wresp.Header.Revision, 10),
			}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}

// Eval evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
func (s *Store) Eval(ctx context.Context, path string, params regula.ParamGetter) (*regula.EvalResult, error) {
	re, err := s.Latest(ctx, path)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, regula.ErrRulesetNotFound
		}

		return nil, err
	}

	v, err := re.Ruleset.Eval(params)
	if err != nil {
		return nil, err
	}

	return &regula.EvalResult{
		Value:   v,
		Version: re.Version,
	}, nil
}

// EvalVersion evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
func (s *Store) EvalVersion(ctx context.Context, path, version string, params regula.ParamGetter) (*regula.EvalResult, error) {
	re, err := s.OneByVersion(ctx, path, version)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, regula.ErrRulesetNotFound
		}

		return nil, err
	}

	v, err := re.Ruleset.Eval(params)
	if err != nil {
		return nil, err
	}

	return &regula.EvalResult{
		Value:   v,
		Version: re.Version,
	}, nil
}

func (s *Store) rulesetPath(p, v string) string {
	return path.Join(s.Namespace, "rulesets", p, v)
}

func (s *Store) checksumPath(p string) string {
	return path.Join(s.Namespace, "checksums", p)
}
