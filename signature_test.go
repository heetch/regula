package regula

import (
	"fmt"
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

// Equal should return an well described Error if the signatures aren't the same.
func TestSignatureEquality(t *testing.T) {
	rs, err := NewBoolRuleset(rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("baz")), rule.BoolValue(true)))
	require.NoError(t, err)
	rootSig := NewSignature(rs)

	t.Run("OK", func(t *testing.T) {
		ok, err := rootSig.Equal(rootSig)
		require.True(t, ok)
		require.Nil(t, err)
	})

	t.Run("Bad return type", func(t *testing.T) {
		rs1, err := NewStringRuleset(rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("baz")), rule.StringValue("true")))
		require.NoError(t, err)
		sig := NewSignature(rs1)

		ok, err := rootSig.Equal(sig)
		require.False(t, ok)
		exp := &Error{
			Field:  "return type",
			Value:  sig.ReturnType,
			Reason: "signature mismatch: return type must be of type bool",
		}
		require.EqualValues(t, exp, err)
		require.Equal(t, fmt.Sprintf("invalid %s with value '%s': %s", exp.Field, exp.Value, exp.Reason), exp.Error())
	})

	t.Run("Bad param type", func(t *testing.T) {
		rs1, err := NewBoolRuleset(rule.New(rule.Eq(rule.BoolParam("foo"), rule.StringValue("baz")), rule.BoolValue(true)))
		require.NoError(t, err)
		sig := NewSignature(rs1)

		ok, err := rootSig.Equal(sig)
		require.False(t, ok)
		exp := &Error{
			Field:  "param type",
			Value:  "bool",
			Reason: "signature mismatch: param must be of type string",
		}
		require.EqualValues(t, exp, err)
		require.Equal(t, fmt.Sprintf("invalid %s with value '%s': %s", exp.Field, exp.Value, exp.Reason), exp.Error())
	})

	t.Run("Bad param", func(t *testing.T) {
		rs1, err := NewBoolRuleset(rule.New(rule.Eq(rule.StringParam("bar"), rule.StringValue("baz")), rule.BoolValue(true)))
		require.NoError(t, err)
		sig := NewSignature(rs1)

		ok, err := rootSig.Equal(sig)
		require.False(t, ok)
		exp := &Error{
			Field:  "param",
			Value:  "bar",
			Reason: "signature mismatch: unknown parameter",
		}
		require.EqualValues(t, exp, err)
		require.Equal(t, fmt.Sprintf("invalid %s with value '%s': %s", exp.Field, exp.Value, exp.Reason), exp.Error())
	})
}
