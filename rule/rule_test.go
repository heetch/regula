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
		require.IsType(t, new(nodeValue), eq.Operands[0])
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
			{Eq(ValueStr("foo"), ValueStr("foo")), nil},
			{Eq(ValueStr("foo"), ParamStr("bar")), Params{"bar": "foo"}},
			{In(ValueStr("foo"), ParamStr("bar")), Params{"bar": "foo"}},
			{
				Eq(
					Eq(ValueStr("bar"), ValueStr("bar")),
					Eq(ValueStr("foo"), ValueStr("foo")),
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
			{ValueStr("foo"), nil},
			{ParamStr("bar"), Params{"bar": "foo"}},
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
		r, err := NewRuleset(
			"string",
			New(Eq(ValueStr("foo"), ValueStr("bar")), ReturnsStr("first")),
			New(Eq(ValueStr("foo"), ValueStr("foo")), ReturnsStr("second")),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "second", res.Value)
	})

	t.Run("Match bool", func(t *testing.T) {
		r, err := NewRuleset(
			"bool",
			New(Eq(ValueStr("foo"), ValueStr("bar")), ReturnsBool(false)),
			New(Eq(ValueStr("foo"), ValueStr("foo")), ReturnsBool(true)),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "true", res.Value)
	})

	t.Run("Type mismatch", func(t *testing.T) {
		_, err := NewRuleset(
			"string",
			New(Eq(ValueStr("foo"), ValueStr("bar")), ReturnsStr("first")),
			New(Eq(ValueStr("foo"), ValueStr("foo")), ReturnsBool(true)),
		)
		require.Equal(t, ErrRulesetIncoherentType, err)
	})

	t.Run("No match", func(t *testing.T) {
		r, err := NewRuleset(
			"string",
			New(Eq(ValueStr("foo"), ValueStr("bar")), ReturnsStr("first")),
			New(Eq(ValueStr("bar"), ValueStr("foo")), ReturnsStr("second")),
		)
		require.NoError(t, err)

		_, err = r.Eval(nil)
		require.Equal(t, ErrNoMatch, err)
	})

	t.Run("Default", func(t *testing.T) {
		r, err := NewRuleset(
			"string",
			New(Eq(ValueStr("foo"), ValueStr("bar")), ReturnsStr("first")),
			New(Eq(ValueStr("bar"), ValueStr("foo")), ReturnsStr("second")),
			New(True(), ReturnsStr("default")),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "default", res.Value)
	})
}
