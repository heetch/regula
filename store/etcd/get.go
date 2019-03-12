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
func (s *RulesetService) Get(ctx context.Context, path, version string) (*regula.Ruleset, error) {
	if path == "" {
		return nil, store.ErrRulesetNotFound
	}

	ops := []clientv3.Op{
		clientv3.OpGet(s.signaturesPath(path)),
		clientv3.OpGet(s.versionsPath(path)),
	}

	if version == "" {
		ops = append(ops, clientv3.OpGet(s.rulesPath(path, "")+versionSeparator, clientv3.WithLastKey()...))
	} else {
		ops = append(ops, clientv3.OpGet(s.rulesPath(path, version)))
	}
	// running all the requests within a single transaction so only one network round trip is performed.
	resp, err := s.Client.KV.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch ruleset: %s", path)
	}

	if len(resp.Responses) != 3 {
		return nil, store.ErrRulesetNotFound
	}

	// decode signature
	var sig pb.Signature
	gresp := resp.Responses[0].GetResponseRange()
	if gresp.Count == 0 {
		return nil, store.ErrRulesetNotFound
	}
	err = proto.Unmarshal(gresp.Kvs[0].Value, &sig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signature")
	}

	// decode versions
	var versions pb.Versions
	err = proto.Unmarshal(resp.Responses[1].GetResponseRange().Kvs[0].Value, &versions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal versions")
	}

	// decode rules
	var rules pb.Rules
	err = proto.Unmarshal(resp.Responses[2].GetResponseRange().Kvs[0].Value, &rules)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ruleset")
	}
	if version == "" {
		_, version = s.pathVersionFromKey(string(resp.Responses[2].GetResponseRange().Kvs[0].Key))
	}

	return &regula.Ruleset{
		Path:      path,
		Version:   version,
		Rules:     rulesFromProtobuf(&rules),
		Signature: signatureFromProtobuf(&sig),
		Versions:  versions.Versions,
	}, nil
}
