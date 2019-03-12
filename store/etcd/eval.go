package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/pkg/errors"
)

// Eval evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
func (s *RulesetService) Eval(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
	var res *clientv3.GetResponse
	var err error

	if version == "" {
		res, err = s.Client.Get(ctx, s.rulesPath(path, "")+versionSeparator, clientv3.WithLastKey()...)
	} else {
		res, err = s.Client.Get(ctx, s.rulesPath(path, version))
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch ruleset: %s", path)
	}

	if res.Count == 0 {
		return nil, rerrors.ErrRulesetNotFound
	}

	var pr pb.Rules
	err = proto.Unmarshal(res.Kvs[0].Value, &pr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ruleset")
	}

	v, err := regula.NewRuleset(rulesFromProtobuf(&pr)...).Eval(params)
	if err != nil {
		return nil, err
	}

	if version == "" {
		_, version = s.pathVersionFromKey(string(res.Kvs[0].Key))
	}

	return &regula.EvalResult{
		Value:   v,
		Version: version,
	}, nil
}
