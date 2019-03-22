package etcd

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/pkg/errors"
)

func computeLimit(l int) int {
	if l <= 0 || l > 100 {
		return 50 // TODO: make this one configurable
	}
	return l
}

// List returns the latest version of each ruleset under the given prefix.
// If the prefix is empty, it returns rulesets from the beginning following the lexical order.
// The listing can be customised using the ListOptions type.
func (s *RulesetService) List(ctx context.Context, prefix string, opt *api.ListOptions) (*api.Rulesets, error) {
	options := make([]clientv3.OpOption, 0, 3)

	var key string

	limit := computeLimit(opt.Limit)

	if opt.ContinueToken != "" {
		lastPath, err := base64.URLEncoding.DecodeString(opt.ContinueToken)
		if err != nil {
			return nil, api.ErrInvalidContinueToken
		}

		key = string(lastPath)

		var rangeEnd string
		if opt.AllVersions {
			rangeEnd = clientv3.GetPrefixRangeEnd(s.rulesPath(prefix, ""))
		} else {
			rangeEnd = clientv3.GetPrefixRangeEnd(s.latestVersionPath(prefix))
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
func (s *RulesetService) listPathsOnly(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*api.Rulesets, error) {
	opts = append(opts, clientv3.WithKeysOnly())
	resp, err := s.Client.KV.Get(ctx, s.latestVersionPath(key), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch versions")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, api.ErrRulesetNotFound
	}

	rulesets := api.Rulesets{
		Rulesets: make([]regula.Ruleset, len(resp.Kvs)),
	}
	rulesets.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	for i, pair := range resp.Kvs {
		p := strings.TrimPrefix(string(pair.Key), s.latestVersionPath("")+"/")
		rulesets.Rulesets[i].Path = p
	}

	if len(rulesets.Rulesets) < limit || !resp.More {
		return &rulesets, nil
	}

	lastEntry := rulesets.Rulesets[len(rulesets.Rulesets)-1]

	// we want to start immediately after the last key
	rulesets.Continue = base64.URLEncoding.EncodeToString([]byte(lastEntry.Path + "\x00"))

	return &rulesets, nil
}

// listLastVersion returns only the latest version for each ruleset.
func (s *RulesetService) listLastVersion(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*api.Rulesets, error) {
	resp, err := s.Client.KV.Get(ctx, s.latestVersionPath(key), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch latest keys")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, api.ErrRulesetNotFound
	}

	ops := make([]clientv3.Op, 0, resp.Count)
	txn := s.Client.KV.Txn(ctx)
	for _, pair := range resp.Kvs {
		ops = append(ops, clientv3.OpGet(string(pair.Value)))
	}
	txnresp, err := txn.Then(ops...).Commit()
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed to fetch selected rulesets")
	}

	var rulesets api.Rulesets
	rulesets.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	rulesets.Rulesets = make([]regula.Ruleset, len(resp.Kvs))

	// Responses handles responses for each OpGet calls in the transaction.
	for i, resps := range txnresp.Responses {
		var pbr pb.Rules
		rr := resps.GetResponseRange()

		// Given that we are getting a leaf in the tree (a ruleset entry), we are sure that we will always have one value in the Kvs slice.
		err = proto.Unmarshal(rr.Kvs[0].Value, &pbr)
		if err != nil {
			s.Logger.Debug().Err(err).Bytes("entry", rr.Kvs[0].Value).Msg("list: unmarshalling failed")
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}

		path, version := s.pathVersionFromKey(string(rr.Kvs[0].Key))
		rulesets.Rulesets[i] = regula.Ruleset{
			Path:    path,
			Version: version,
			Rules:   rulesFromProtobuf(&pbr),
		}
	}

	if len(rulesets.Rulesets) < limit || !resp.More {
		return &rulesets, nil
	}

	lastEntry := rulesets.Rulesets[len(rulesets.Rulesets)-1]

	// we want to start immediately after the last key
	rulesets.Continue = base64.URLEncoding.EncodeToString([]byte(lastEntry.Path + "\x00"))

	return &rulesets, nil
}

// listAllVersions returns all available versions for each ruleset.
func (s *RulesetService) listAllVersions(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*api.Rulesets, error) {
	resp, err := s.Client.KV.Get(ctx, s.rulesPath(key, ""), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all rulesets")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, api.ErrRulesetNotFound
	}

	var rulesets api.Rulesets
	rulesets.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	rulesets.Rulesets = make([]regula.Ruleset, len(resp.Kvs))
	for i, pair := range resp.Kvs {
		var pbr pb.Rules

		err = proto.Unmarshal(pair.Value, &pbr)
		if err != nil {
			s.Logger.Debug().Err(err).Bytes("entry", pair.Value).Msg("list: unmarshalling failed")
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}

		path, version := s.pathVersionFromKey(string(pair.Key))
		rulesets.Rulesets[i] = regula.Ruleset{
			Path:    path,
			Version: version,
			Rules:   rulesFromProtobuf(&pbr),
		}
	}

	if len(rulesets.Rulesets) < limit || !resp.More {
		return &rulesets, nil
	}

	lastEntry := rulesets.Rulesets[len(rulesets.Rulesets)-1]

	// we want to start immediately after the last key
	rulesets.Continue = base64.URLEncoding.EncodeToString([]byte(lastEntry.Path + versionSeparator + lastEntry.Version + "\x00"))

	return &rulesets, nil
}
