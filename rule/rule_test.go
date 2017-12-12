package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleUnmarshalling(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var rule Rule

		err := rule.UnmarshalJSON([]byte(`{
			"result": {
				"value": "foo",
                "type": "string"
			},
			"root": {
				"kind": "eq",
				"operands": [
					{
						"kind": "value",
						"type": "string",
						"value": "bar"
					},
					{
						"kind": "eq",
						"operands": [
							{
								"kind": "param",
								"type": "string",
								"name": "foo"
							},
							{
								"kind": "value",
								"type": "string",
								"value": "bar"
							}
						]
					}
				]
			}
		}`))
		require.NoError(t, err)
		require.Equal(t, "string", rule.Result.Type)
		require.Equal(t, "foo", rule.Result.Value)
		require.IsType(t, new(nodeEq), rule.Root)
		eq := rule.Root.(*nodeEq)
		require.Len(t, eq.Operands, 2)
		require.IsType(t, new(Value), eq.Operands[0])
		require.IsType(t, new(nodeEq), eq.Operands[1])
	})

	t.Run("Missing result type", func(t *testing.T) {
		var rule Rule

		err := rule.UnmarshalJSON([]byte(`{
			"result": {
				"value": "foo"
			},
			"root": {
				"kind": "not",
				"operands": [
					{
						"kind": "value",
						"type": "bool",
						"value": "true"
					}
				]
			}
		}`))

		require.Error(t, err)
	})
}

func TestRuleEval(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		tests := []struct {
			node   Node
			params Params
		}{
			{Eq(StringValue("foo"), StringValue("foo")), nil},
			{Eq(StringValue("foo"), StringParam("bar")), Params{"bar": "foo"}},
			{In(StringValue("foo"), StringParam("bar")), Params{"bar": "foo"}},
			{
				Eq(
					Eq(StringValue("bar"), StringValue("bar")),
					Eq(StringValue("foo"), StringValue("foo")),
				),
				nil,
			},
			{True(), nil},
		}

		for _, test := range tests {
			r := New(test.node, ReturnsStr("matched"))
			res, err := r.Eval(test.params)
			require.NoError(t, err)
			require.Equal(t, "matched", res.Value)
			require.Equal(t, "string", res.Type)
		}
	})

	t.Run("Invalid return", func(t *testing.T) {
		tests := []struct {
			node   Node
			params Params
		}{
			{StringValue("foo"), nil},
			{StringParam("bar"), Params{"bar": "foo"}},
		}

		for _, test := range tests {
			r := New(test.node, ReturnsStr("matched"))
			_, err := r.Eval(test.params)
			require.Error(t, err)
		}
	})
}

func TestRulesetEval(t *testing.T) {
	t.Run("Match string", func(t *testing.T) {
		r, err := NewStringRuleset(
			New(Eq(StringValue("foo"), StringValue("bar")), ReturnsStr("first")),
			New(Eq(StringValue("foo"), StringValue("foo")), ReturnsStr("second")),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "second", res.Value)
	})

	t.Run("Match bool", func(t *testing.T) {
		r, err := NewBoolRuleset(
			New(Eq(StringValue("foo"), StringValue("bar")), ReturnsBool(false)),
			New(Eq(StringValue("foo"), StringValue("foo")), ReturnsBool(true)),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "true", res.Value)
	})

	t.Run("Type mismatch", func(t *testing.T) {
		_, err := NewStringRuleset(
			New(Eq(StringValue("foo"), StringValue("bar")), ReturnsStr("first")),
			New(Eq(StringValue("foo"), StringValue("foo")), ReturnsBool(true)),
		)
		require.Equal(t, ErrRulesetIncoherentType, err)
	})

	t.Run("No match", func(t *testing.T) {
		r, err := NewStringRuleset(
			New(Eq(StringValue("foo"), StringValue("bar")), ReturnsStr("first")),
			New(Eq(StringValue("bar"), StringValue("foo")), ReturnsStr("second")),
		)
		require.NoError(t, err)

		_, err = r.Eval(nil)
		require.Equal(t, ErrNoMatch, err)
	})

	t.Run("Default", func(t *testing.T) {
		r, err := NewStringRuleset(
			New(Eq(StringValue("foo"), StringValue("bar")), ReturnsStr("first")),
			New(Eq(StringValue("bar"), StringValue("foo")), ReturnsStr("second")),
			New(True(), ReturnsStr("default")),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "default", res.Value)
	})
}
