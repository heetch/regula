package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/pkg/errors"
)

// Get returns the ruleset related to the given path.
func (s *RulesetService) Get(ctx context.Context, path string) (*regula.Ruleset, error) {
	if path == "" {
		return nil, api.ErrRulesetNotFound
	}

	ops := []clientv3.Op{
		clientv3.OpGet(s.signaturesPath(path)),
		clientv3.OpGet(s.rulesPath(path, "")+versionSeparator, clientv3.WithLimit(100), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend)),
	}

	// running all the requests within a single transaction so only one network round trip is performed.
	resp, err := s.Client.KV.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch ruleset: %s", path)
	}

	if len(resp.Responses) != 2 {
		return nil, api.ErrRulesetNotFound
	}

	// decode signature
	var sig pb.Signature
	gresp := resp.Responses[0].GetResponseRange()
	if gresp.Count == 0 {
		return nil, api.ErrRulesetNotFound
	}
	err = proto.Unmarshal(gresp.Kvs[0].Value, &sig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signature")
	}

	// decode versions of rules. they are returned by etcd ordered by key descendend.
	// this loop rearranges them in ascending order.
	versions := make([]regula.RulesetVersion, len(resp.Responses[1].GetResponseRange().Kvs))
	for i, kv := range resp.Responses[1].GetResponseRange().Kvs {
		var rules pb.Rules
		err = proto.Unmarshal(kv.Value, &rules)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal ruleset")
		}

		versions[len(versions)-i].Rules = rulesFromProtobuf(&rules)
		_, versions[len(versions)-i].Version = s.pathVersionFromKey(string(kv.Key))
	}

	return &regula.Ruleset{
		Path:      path,
		Signature: signatureFromProtobuf(&sig),
		Versions:  versions,
	}, nil
}
