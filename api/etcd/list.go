package etcd

import (
	"context"
	"encoding/base64"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/pkg/errors"
)

// List returns rulesets whose path starts by the given prefix.
// The listing is paginated and can be customised using the ListOptions type.
// It runs two requests, one for fetching the signatures and one for fetching the related rules versions.
func (s *RulesetService) List(ctx context.Context, prefix string, opt *api.ListOptions) (*api.Rulesets, error) {
	opts := make([]clientv3.OpOption, 0, 3)

	if opt == nil {
		opt = new(api.ListOptions)
	}

	var key string

	// if a cursor is specified, decode the key from it and start the request from that key
	if opt.Cursor != "" {
		lastPath, err := base64.URLEncoding.DecodeString(opt.Cursor)
		if err != nil {
			return nil, api.ErrInvalidCursor
		}

		key = string(lastPath)

		opts = append(opts, clientv3.WithRange(clientv3.GetPrefixRangeEnd(s.rulesPath(prefix, ""))))
	} else {
		key = prefix
		opts = append(opts, clientv3.WithPrefix())
	}

	// limit the number of results
	opts = append(opts, clientv3.WithLimit(int64(opt.GetLimit())))

	// fetch signatures
	resp, err := s.Client.KV.Get(ctx, s.signaturesPath(key), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch signatures")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, api.ErrRulesetNotFound
	}

	var ops []clientv3.Op

	rulesets := api.Rulesets{
		Revision: strconv.FormatInt(resp.Header.Revision, 10),
		Rulesets: make([]regula.Ruleset, len(resp.Kvs)),
	}

	for i, pair := range resp.Kvs {
		var sig pb.Signature

		err = proto.Unmarshal(pair.Value, &sig)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal signature")
		}

		rulesets.Rulesets[i].Path = s.pathFromKey("signatures", string(pair.Key))
		rulesets.Rulesets[i].Signature = signatureFromProtobuf(&sig)

		if opt.LatestVersionsLimit > 0 {
			ops = append(ops, clientv3.OpGet(s.rulesPath(rulesets.Rulesets[i].Path, ""), clientv3.WithLimit(int64(opt.LatestVersionsLimit))))
		}
	}

	if len(ops) > 0 {
		err = s.fetchRulesVersions(ctx, &rulesets, ops)
		if err != nil {
			return nil, err
		}
	}

	if len(rulesets.Rulesets) < opt.GetLimit() || !resp.More {
		return &rulesets, nil
	}

	lastRuleset := rulesets.Rulesets[len(rulesets.Rulesets)-1]

	// we want to start immediately after the last key
	rulesets.Cursor = base64.URLEncoding.EncodeToString([]byte(lastRuleset.Path + "\x00"))

	return &rulesets, nil
}

func (s *RulesetService) fetchRulesVersions(ctx context.Context, rulesets *api.Rulesets, ops []clientv3.Op) error {
	// running all the requests within a single transaction so only one network round trip is performed.
	rulesResp, err := s.Client.KV.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return errors.Wrapf(err, "failed to fetch rules")
	}

	for i, rop := range rulesResp.Responses {
		respRange := rop.GetResponseRange()
		if respRange.Count == 0 {
			continue
		}

		rulesets.Rulesets[i].Versions = make([]regula.RulesetVersion, int(respRange.Count))
		for j, kv := range respRange.Kvs {
			var pbr pb.Rules
			err = proto.Unmarshal(kv.Value, &pbr)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal rules")
			}

			_, rulesets.Rulesets[i].Versions[j].Version = s.pathVersionFromKey(string(kv.Key))
			rulesets.Rulesets[i].Versions[j].Rules = rulesFromProtobuf(&pbr)
		}
	}

	return nil
}
