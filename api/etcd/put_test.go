package etcd

import (
	"context"
	ppath "path"
	"strings"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestPut(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	path := "a"
	sig := &regula.Signature{ReturnType: "bool"}
	require.NoError(t, s.Create(context.Background(), path, sig))

	t.Run("OK", func(t *testing.T) {
		path := "a"
		rules := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}

		// put the rules
		version, err := s.Put(context.Background(), path, rules)
		require.NoError(t, err)
		require.NotEmpty(t, version)

		// verify rules creation
		resp, err := s.Client.Get(context.Background(), s.rulesPath(path, ""), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)
		var pbrules pb.Rules
		err = proto.Unmarshal(resp.Kvs[0].Value, &pbrules)
		require.EqualValues(t, rules, rulesFromProtobuf(&pbrules))

		// verify if the path contains the right rules version
		require.Equal(t, version, strings.TrimPrefix(string(resp.Kvs[0].Key), ppath.Join(s.Namespace, "rulesets", "rules", "a")+"!"))

		// verify checksum creation
		resp, err = s.Client.Get(context.Background(), s.checksumsPath(path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		// create new version with same ruleset
		newVersion, err := s.Put(context.Background(), path, rules)
		require.Equal(t, api.ErrRulesetNotModified, err)
		require.Equal(t, version, newVersion)

		// create new version with different rules
		rules = []*rule.Rule{rule.New(rule.True(), rule.BoolValue(false))}

		newVersion, err = s.Put(context.Background(), path, rules)
		require.NoError(t, err)
		require.NotEqual(t, version, newVersion)

		// verify new rules creation
		resp, err = s.Client.Get(context.Background(), s.rulesPath(path, newVersion), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)
		err = proto.Unmarshal(resp.Kvs[0].Value, &pbrules)
		require.NoError(t, err)
		require.EqualValues(t, rules, rulesFromProtobuf(&pbrules))
	})

	t.Run("Signature checks", func(t *testing.T) {
		path := "b"
		require.NoError(t, s.Create(context.Background(), path, &regula.Signature{ReturnType: "bool", Params: map[string]string{"a": "string", "b": "bool", "c": "int64"}}))

		rules1 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
				),
				rule.BoolValue(true),
			),
		}

		_, err := s.Put(context.Background(), path, rules1)
		require.NoError(t, err)

		// same params, different return type
		rules2 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
				),
				rule.StringValue("true"),
			),
		}

		_, err = s.Put(context.Background(), path, rules2)
		require.True(t, api.IsValidationError(err))

		// adding new params
		rs3 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		}

		_, err = s.Put(context.Background(), path, rs3)
		require.True(t, api.IsValidationError(err))

		// changing param types
		rs4 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.StringParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		}

		_, err = s.Put(context.Background(), path, rs4)
		require.True(t, api.IsValidationError(err))

		// adding new rule with different param types
		rs5 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.StringParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.StringParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		}

		_, err = s.Put(context.Background(), path, rs5)
		require.True(t, api.IsValidationError(err))

		// adding new rule with correct param types but less
		rs6 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
				),
				rule.BoolValue(true),
			),
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
				),
				rule.BoolValue(true),
			),
		}

		_, err = s.Put(context.Background(), path, rs6)
		require.NoError(t, err)
	})
}
