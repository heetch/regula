package etcd

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
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

	exp := "test/rulesets/rulesets/path" + versionSeparator + "version"
	require.Equal(t, exp, s.rulesetsPath("path", "version"))

	exp = "test/rulesets/rulesets/path"
	require.Equal(t, exp, s.rulesetsPath("path", ""))

	exp = "test/rulesets/checksums/path"
	require.Equal(t, exp, s.checksumsPath("path"))

	exp = "test/rulesets/signatures/path"
	require.Equal(t, exp, s.signaturesPath("path"))

	exp = "test/rulesets/latest/path"
	require.Equal(t, exp, s.latestVersionPath("path"))

	exp = "test/rulesets/versions/path"
	require.Equal(t, exp, s.versionsPath("path"))
}

// compareSignature should return a ValidationError if the signatures aren't the same.
func TestCompareSignature(t *testing.T) {
	rs := regula.NewRuleset(rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("baz")), rule.BoolValue(true)))
	baseSig := regula.NewSignature(rs)

	t.Run("OK", func(t *testing.T) {
		err := compareSignature(baseSig, baseSig)
		require.Nil(t, err)
	})

	t.Run("Bad return type", func(t *testing.T) {
		rs1, err := regula.NewStringRuleset(rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("baz")), rule.StringValue("true")))
		require.NoError(t, err)
		sig := regula.NewSignature(rs1)

		err = compareSignature(baseSig, sig)
		exp := &store.ValidationError{
			Field:  "return type",
			Value:  sig.ReturnType,
			Reason: "signature mismatch: return type must be of type bool",
		}
		require.EqualValues(t, exp, err)
		require.Equal(t, fmt.Sprintf("invalid %s with value '%s': %s", exp.Field, exp.Value, exp.Reason), exp.Error())
	})

	t.Run("Bad param type", func(t *testing.T) {
		rs1, err := regula.NewRuleset(rule.New(rule.Eq(rule.BoolParam("foo"), rule.StringValue("baz")), rule.BoolValue(true)))
		require.NoError(t, err)
		sig := regula.NewSignature(rs1)

		err = compareSignature(baseSig, sig)
		exp := &store.ValidationError{
			Field:  "param type",
			Value:  "bool",
			Reason: "signature mismatch: param must be of type string",
		}
		require.EqualValues(t, exp, err)
		require.Equal(t, fmt.Sprintf("invalid %s with value '%s': %s", exp.Field, exp.Value, exp.Reason), exp.Error())
	})

	t.Run("Bad param", func(t *testing.T) {
		rs1, err := regula.NewRuleset(rule.New(rule.Eq(rule.StringParam("bar"), rule.StringValue("baz")), rule.BoolValue(true)))
		require.NoError(t, err)
		sig := regula.NewSignature(rs1)

		err = compareSignature(baseSig, sig)
		exp := &store.ValidationError{
			Field:  "param",
			Value:  "bar",
			Reason: "signature mismatch: unknown parameter",
		}
		require.EqualValues(t, exp, err)
		require.Equal(t, fmt.Sprintf("invalid %s with value '%s': %s", exp.Field, exp.Value, exp.Reason), exp.Error())
	})
}

func BenchmarkProtoMarshalling(b *testing.B) {
	rs := regula.NewRuleset(
		rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.BoolValue(true)),
		rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.BoolValue(false)),
	)
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := proto.Marshal(rulesetToProtobuf(rs))
		require.NoError(b, err)
	}
}

func BenchmarkJSONMarshalling(b *testing.B) {
	rs := regula.NewRuleset(
		rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.BoolValue(true)),
		rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.BoolValue(false)),
	)
	require.NoError(b, err)

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
	require.NoError(b, err)

	bb, err := proto.Marshal(rulesetToProtobuf(rs))
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var pbrs pb.Ruleset
		err := proto.Unmarshal(bb, &pbrs)
		require.NoError(b, err)
	}
}

func BenchmarkJSONUnmarshalling(b *testing.B) {
	rs := regula.NewRuleset(
		rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.BoolValue(true)),
		rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.BoolValue(false)),
	)
	require.NoError(b, err)

	bb, err := json.Marshal(rs)
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var rrs regula.Ruleset
		err := json.Unmarshal(bb, &rrs)
		require.NoError(b, err)
	}
}
