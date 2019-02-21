package etcd

import (
	"context"
	"encoding/base64"
	"path"
	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula/store"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// versionSeparator separates the path from the version in the entries path in etcd.
// The purpose is to have the same ordering as the others namespace (latest, versions, ...).
const versionSeparator = "!"

// RulesetService manages the rulesets using etcd.
type RulesetService struct {
	Client    *clientv3.Client
	Logger    zerolog.Logger
	Namespace string
}

func computeLimit(l int) int {
	if l <= 0 || l > 100 {
		return 50 // TODO: make this one configurable
	}
	return l
}

// List returns the latest version of each ruleset under the given prefix.
// If the prefix is empty, it returns entries from the beginning following the lexical order.
// The listing can be customised using the ListOptions type.
func (s *RulesetService) List(ctx context.Context, prefix string, opt *store.ListOptions) (*store.RulesetEntries, error) {
	options := make([]clientv3.OpOption, 0, 3)

	var key string

	limit := computeLimit(opt.Limit)

	if opt.ContinueToken != "" {
		lastPath, err := base64.URLEncoding.DecodeString(opt.ContinueToken)
		if err != nil {
			return nil, store.ErrInvalidContinueToken
		}

		key = string(lastPath)

		var rangeEnd string
		if opt.AllVersions {
			rangeEnd = clientv3.GetPrefixRangeEnd(s.entriesPath(prefix, ""))
		} else {
			rangeEnd = clientv3.GetPrefixRangeEnd(s.latestRulesetPath(prefix))
		}
		options = append(options, clientv3.WithRange(rangeEnd))
	} else {
		key = prefix
		options = append(options, clientv3.WithPrefix())
	}

	options = append(options, clientv3.WithLimit(int64(limit)))

	switch {
	case opt.PathsOnly:
		return s.listPathsOnly(ctx, key, prefix, limit, options)
	case opt.AllVersions:
		return s.listAllVersions(ctx, key, prefix, limit, options)
	default:
		return s.listLastVersion(ctx, key, prefix, limit, options)
	}
}

// listPathsOnly returns only the path for each ruleset.
func (s *RulesetService) listPathsOnly(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*store.RulesetEntries, error) {
	opts = append(opts, clientv3.WithKeysOnly())
	resp, err := s.Client.KV.Get(ctx, s.latestRulesetPath(key), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all entries")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, store.ErrNotFound
	}

	var entries store.RulesetEntries
	entries.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	for _, pair := range resp.Kvs {
		p := strings.TrimPrefix(string(pair.Key), s.latestRulesetPath("")+"/")
		entries.Entries = append(entries.Entries, store.RulesetEntry{Path: p})
	}

	if len(entries.Entries) < limit || !resp.More {
		return &entries, nil
	}

	lastEntry := entries.Entries[len(entries.Entries)-1]

	// we want to start immediately after the last key
	entries.Continue = base64.URLEncoding.EncodeToString([]byte(lastEntry.Path + "\x00"))

	return &entries, nil
}

// listLastVersion returns only the latest version for each ruleset.
func (s *RulesetService) listLastVersion(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*store.RulesetEntries, error) {
	resp, err := s.Client.KV.Get(ctx, s.latestRulesetPath(key), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch latests keys")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, store.ErrNotFound
	}

	ops := make([]clientv3.Op, 0, resp.Count)
	txn := s.Client.KV.Txn(ctx)
	for _, pair := range resp.Kvs {
		ops = append(ops, clientv3.OpGet(string(pair.Value)))
	}
	txnresp, err := txn.Then(ops...).Commit()
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed to fetch all entries")
	}

	var entries store.RulesetEntries
	entries.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	entries.Entries = make([]store.RulesetEntry, len(resp.Kvs))

	// Responses handles responses for each OpGet calls in the transaction.
	for i, resps := range txnresp.Responses {
		var pbrse pb.RulesetEntry
		rr := resps.GetResponseRange()

		// Given that we are getting a leaf in the tree (a ruleset entry), we are sure that we will always have one value in the Kvs slice.
		err = proto.Unmarshal(rr.Kvs[0].Value, &pbrse)
		if err != nil {
			s.Logger.Debug().Err(err).Bytes("entry", rr.Kvs[0].Value).Msg("list: unmarshalling failed")
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}

		entries.Entries[i] = store.RulesetEntry{
			Path:    pbrse.Path,
			Version: pbrse.Version,
			Ruleset: rulesetFromProtobuf(pbrse.Ruleset),
		}
	}

	if len(entries.Entries) < limit || !resp.More {
		return &entries, nil
	}

	lastEntry := entries.Entries[len(entries.Entries)-1]

	// we want to start immediately after the last key
	entries.Continue = base64.URLEncoding.EncodeToString([]byte(lastEntry.Path + "\x00"))

	return &entries, nil
}

// listAllVersions returns all available versions for each ruleset.
func (s *RulesetService) listAllVersions(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*store.RulesetEntries, error) {
	resp, err := s.Client.KV.Get(ctx, s.entriesPath(key, ""), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all entries")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, store.ErrNotFound
	}

	var entries store.RulesetEntries
	entries.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	entries.Entries = make([]store.RulesetEntry, len(resp.Kvs))
	for i, pair := range resp.Kvs {
		var pbrse pb.RulesetEntry

		err = proto.Unmarshal(pair.Value, &pbrse)
		if err != nil {
			s.Logger.Debug().Err(err).Bytes("entry", pair.Value).Msg("list: unmarshalling failed")
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}

		entries.Entries[i] = store.RulesetEntry{
			Path:    pbrse.Path,
			Version: pbrse.Version,
			Ruleset: rulesetFromProtobuf(pbrse.Ruleset),
		}
	}

	if len(entries.Entries) < limit || !resp.More {
		return &entries, nil
	}

	lastEntry := entries.Entries[len(entries.Entries)-1]

	// we want to start immediately after the last key
	entries.Continue = base64.URLEncoding.EncodeToString([]byte(lastEntry.Path + versionSeparator + lastEntry.Version + "\x00"))

	return &entries, nil
}

// Watch the given prefix for anything new.
func (s *RulesetService) Watch(ctx context.Context, prefix string, revision string) (*store.RulesetEvents, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	if i, _ := strconv.ParseInt(revision, 10, 64); i > 0 {
		// watch from the next revision
		opts = append(opts, clientv3.WithRev(i+1))
	}

	wc := s.Client.Watch(ctx, s.entriesPath(prefix, ""), opts...)
	for {
		select {
		case wresp := <-wc:
			if err := wresp.Err(); err != nil {
				return nil, errors.Wrapf(err, "failed to watch prefix: '%s'", prefix)
			}

			if len(wresp.Events) == 0 {
				continue
			}

			events := make([]store.RulesetEvent, len(wresp.Events))
			for i, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					events[i].Type = store.RulesetPutEvent
				default:
					s.Logger.Debug().Str("type", string(ev.Type)).Msg("watch: ignoring event type")
					continue
				}

				var pbrse pb.RulesetEntry
				err := proto.Unmarshal(ev.Kv.Value, &pbrse)
				if err != nil {
					s.Logger.Debug().Bytes("entry", ev.Kv.Value).Msg("watch: unmarshalling failed")
					return nil, errors.Wrap(err, "failed to unmarshal entry")
				}
				events[i].Path = pbrse.Path
				events[i].Ruleset = rulesetFromProtobuf(pbrse.Ruleset)
				events[i].Version = pbrse.Version
			}

			return &store.RulesetEvents{
				Events:   events,
				Revision: strconv.FormatInt(wresp.Header.Revision, 10),
			}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}

// entriesPath returns the path where the rulesets are stored in etcd.
func (s *RulesetService) entriesPath(p, v string) string {
	// If the version parameter is not empty, we concatenate to the path separated by the versionSeparator value.
	if v != "" {
		p += versionSeparator + v
	}
	return path.Join(s.Namespace, "rulesets", "entries", p)
}

// checksumsPath returns the path where the checksums are stored in etcd.
func (s *RulesetService) checksumsPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "checksums", p)
}

// signaturesPath returns the path where the signatures are stored in etcd.
func (s *RulesetService) signaturesPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "signatures", p)
}

// latestRulesetPath returns the path where the latest version of each ruleset is stored in etcd.
func (s *RulesetService) latestRulesetPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "latest", p)
}

// versionsPath returns the path where the versions of each rulesets are stored in etcd.
func (s *RulesetService) versionsPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "versions", p)
}
