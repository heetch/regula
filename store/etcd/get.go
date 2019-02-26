package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/store"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/pkg/errors"
)

// Get returns the ruleset related to the given path. By default, it returns the latest one.
// It returns the related ruleset version if it's specified.
func (s *RulesetService) Get(ctx context.Context, path, version string) (*store.RulesetEntry, error) {
	var (
		entry *store.RulesetEntry
		err   error
	)

	if version == "" {
		entry, err = s.Latest(ctx, path)
	} else {
		entry, err = s.OneByVersion(ctx, path, version)
	}
	if err != nil {
		return nil, err
	}

	resp, err := s.Client.KV.Get(ctx, s.versionsPath(path))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch versions of the entry: %s", path)
	}
	if resp.Count == 0 {
		s.Logger.Debug().Str("path", path).Msg("cannot find ruleset versions list")
		return nil, store.ErrNotFound
	}

	var v pb.Versions
	err = proto.Unmarshal(resp.Kvs[0].Value, &v)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal versions")
	}
	entry.Versions = v.Versions

	resp, err = s.Client.KV.Get(ctx, s.signaturesPath(path))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch signature of the entry: %s", path)
	}
	if resp.Count == 0 {
		s.Logger.Debug().Str("path", path).Msg("cannot find ruleset signature")
		return nil, store.ErrNotFound
	}

	var sig pb.Signature
	err = proto.Unmarshal(resp.Kvs[0].Value, &sig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal signature")
	}
	entry.Signature = signatureFromProtobuf(&sig)

	return entry, nil
}

// Latest returns the latest version of the ruleset entry which corresponds to the given path.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *RulesetService) Latest(ctx context.Context, path string) (*store.RulesetEntry, error) {
	if path == "" {
		return nil, store.ErrNotFound
	}

	// running both requests within a single transaction so only one network round trip is performed.
	resp, err := s.Client.KV.Txn(ctx).
		Then(
			clientv3.OpGet(s.rulesPath(path, "")+versionSeparator, clientv3.WithLastKey()...),
			clientv3.OpGet(s.signaturesPath(path)),
			clientv3.OpGet(s.latestVersionPath(path)),
		).Commit()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch ruleset: %s", path)
	}

	if len(resp.Responses) != 3 {
		return nil, store.ErrNotFound
	}

	// decode rules
	var rules pb.Rules
	err = proto.Unmarshal(resp.Responses[0].GetResponseRange().Kvs[0].Value, &rules)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal rules")
	}

	// decode signature
	var sig pb.Signature
	err = proto.Unmarshal(resp.Responses[1].GetResponseRange().Kvs[0].Value, &sig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signature")
	}

	return &store.RulesetEntry{
		Path:    path,
		Version: string(resp.Responses[2].GetResponseRange().Kvs[0].Value), // decode version
		Ruleset: &regula.Ruleset{
			Signature: signatureFromProtobuf(&sig),
			Rules:     rulesFromProtobuf(&rules),
		},
	}, nil
}

// OneByVersion returns the ruleset entry which corresponds to the given path at the given version.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *RulesetService) OneByVersion(ctx context.Context, path, version string) (*store.RulesetEntry, error) {
	if path == "" {
		return nil, store.ErrNotFound
	}

	resp, err := s.Client.KV.Get(ctx, s.entriesPath(path, version))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch the entry: %s", path)
	}

	// Count will be 0 if the path doesn't exist or if it's not a ruleset.
	if resp.Count == 0 {
		return nil, store.ErrNotFound
	}

	var entry pb.RulesetEntry
	err = proto.Unmarshal(resp.Kvs[0].Value, &entry)
	if err != nil {
		s.Logger.Debug().Err(err).Bytes("entry", resp.Kvs[0].Value).Msg("one-by-version: unmarshalling failed")
		return nil, errors.Wrap(err, "failed to unmarshal entry")
	}

	return &store.RulesetEntry{
		Path:    entry.Path,
		Version: entry.Version,
		Ruleset: rulesetFromProtobuf(entry.Ruleset),
	}, nil
}
