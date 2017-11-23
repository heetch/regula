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
				"value": "foo"
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
								"kind": "variable",
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
		require.Equal(t, "foo", rule.Result.Value)
		require.IsType(t, new(NodeEq), rule.Root)
		eq := rule.Root.(*NodeEq)
		require.Len(t, eq.Operands, 2)
		require.IsType(t, new(NodeValue), eq.Operands[0])
		require.IsType(t, new(NodeEq), eq.Operands[1])
	})
}

func TestRuleEval(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		tests := []struct {
			node Node
			ctx  map[string]string
		}{
			{Eq(ValStr("foo"), ValStr("foo")), nil},
			{Eq(ValStr("foo"), VarStr("bar")), map[string]string{"bar": "foo"}},
			{In(ValStr("foo"), VarStr("bar")), map[string]string{"bar": "foo"}},
			{
				Eq(
					Eq(ValStr("bar"), ValStr("bar")),
					Eq(ValStr("foo"), ValStr("foo")),
				),
				nil,
			},
			{True(), nil},
		}

		for _, test := range tests {
			r := New(test.node, ReturnsStr("matched"))
			res, err := r.Eval(test.ctx)
			require.NoError(t, err)
			require.Equal(t, "matched", res.Value)
			require.Equal(t, "string", res.Type)
		}
	})

	t.Run("Invalid return", func(t *testing.T) {
		tests := []struct {
			node Node
			ctx  map[string]string
		}{
			{ValStr("foo"), nil},
			{VarStr("bar"), map[string]string{"bar": "foo"}},
		}

		for _, test := range tests {
			r := New(test.node, ReturnsStr("matched"))
			_, err := r.Eval(test.ctx)
			require.Error(t, err)
		}
	})
}
