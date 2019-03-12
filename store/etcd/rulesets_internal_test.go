package etcd

import (
	"encoding/json"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/stretchr/testify/require"
)

// Limit should be set to 50 if the given one is <= 0 or > 100.
func TestComputeLimit(t *testing.T) {
	l := computeLimit(0)
	require.Equal(t, 50, l)
	l = computeLimit(-10)
	require.Equal(t, 50, l)
	l = computeLimit(110)
	require.Equal(t, 50, l)
	l = computeLimit(70)
	require.Equal(t, 70, l)
}

// TestPathMethods ensures that the correct path are returned by each method.
func TestPathMethods(t *testing.T) {
	s := &RulesetService{
		Namespace: "test",
	}

	exp := "test/rulesets/rules/path" + versionSeparator + "version"
	require.Equal(t, exp, s.rulesPath("path", "version"))

	exp = "test/rulesets/rules/path"
	require.Equal(t, exp, s.rulesPath("path", ""))

	exp = "test/rulesets/checksums/path"
	require.Equal(t, exp, s.checksumsPath("path"))

	exp = "test/rulesets/signatures/path"
	require.Equal(t, exp, s.signaturesPath("path"))

	exp = "test/rulesets/latest/path"
	require.Equal(t, exp, s.latestVersionPath("path"))

	exp = "test/rulesets/versions/path"
	require.Equal(t, exp, s.versionsPath("path"))
}

func BenchmarkProtoMarshalling(b *testing.B) {
	rs := regula.NewRuleset(
		rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.BoolValue(true)),
		rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.BoolValue(false)),
	)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := proto.Marshal(rulesToProtobuf(rs.Rules))
		require.NoError(b, err)
	}
}

func BenchmarkJSONMarshalling(b *testing.B) {
	rs := regula.NewRuleset(
		rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.BoolValue(true)),
		rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.BoolValue(false)),
	)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := json.Marshal(rs)
		require.NoError(b, err)
	}
}

func BenchmarkProtoUnmarshalling(b *testing.B) {
	rs := regula.NewRuleset(
		rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.BoolValue(true)),
		rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.BoolValue(false)),
	)

	bb, err := proto.Marshal(rulesToProtobuf(rs.Rules))
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var pbrs pb.Rules
		err := proto.Unmarshal(bb, &pbrs)
		require.NoError(b, err)
	}
}

func BenchmarkJSONUnmarshalling(b *testing.B) {
	rs := regula.NewRuleset(
		rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.BoolValue(true)),
		rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.BoolValue(false)),
	)

	bb, err := json.Marshal(rs)
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var rrs regula.Ruleset
		err := json.Unmarshal(bb, &rrs)
		require.NoError(b, err)
	}
}
