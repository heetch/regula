package rule

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperands(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var ops operands

		err := ops.UnmarshalJSON([]byte(`[]`))
		require.NoError(t, err)
		require.Empty(t, ops.Ops)
		require.Empty(t, ops.Nodes)
	})

	t.Run("Some Ops", func(t *testing.T) {
		var ops operands

		err := ops.UnmarshalJSON([]byte(`[
			{"kind": "value"},
			{"kind": "variable"},
			{"kind": "true"},
			{"kind": "eq","operands": [{"kind": "value"}, {"kind": "variable"}]},
			{"kind": "in","operands": [{"kind": "value"}, {"kind": "variable"}]}
		]`))
		require.NoError(t, err)
		require.Len(t, ops.Ops, 5)
		require.Len(t, ops.Nodes, 5)
	})
}

func TestParseNode(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		_, err := parseNode("", []byte(``))
		require.Error(t, err)
	})

	t.Run("Unknown kind", func(t *testing.T) {
		_, err := parseNode("kiwi", []byte(``))
		require.Error(t, err)
	})

	t.Run("OK", func(t *testing.T) {
		tests := []struct {
			kind string
			data []byte
			typ  interface{}
		}{
			{"eq", []byte(`{"kind": "eq","operands": [{"kind": "value"}, {"kind": "variable"}]}`), new(NodeEq)},
			{"in", []byte(`{"kind":"in","operands": [{"kind": "value"}, {"kind": "variable"}]}`), new(NodeIn)},
			{"variable", []byte(`{"kind":"variable"}`), new(NodeVariable)},
			{"value", []byte(`{"kind":"value"}`), new(NodeValue)},
			{"true", []byte(`{"kind":"true"}`), new(NodeTrue)},
		}

		for _, test := range tests {
			n, err := parseNode(test.kind, test.data)
			require.NoError(t, err)
			require.NotNil(t, n)
			require.IsType(t, test.typ, n)
		}
	})
}

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
						"kind": "variable",
						"type": "string",
						"name": "foo"
					},
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
		require.Len(t, eq.Operands, 3)
		require.IsType(t, new(NodeVariable), eq.Operands[0])
		require.IsType(t, new(NodeValue), eq.Operands[1])
		require.IsType(t, new(NodeEq), eq.Operands[2])
	})
}

func TestRuleMarshalling(t *testing.T) {
	t.Run("Eq", func(t *testing.T) {
		r, err := New(
			Eq(VarStr("foo"), ValStr("bar")),
			ReturnsStr("value"),
		)
		require.NoError(t, err)

		raw, err := json.Marshal(r)
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"root": {
					"kind": "eq",
					"operands": [
						{"kind": "variable", "type": "string", "name": "foo"},
						{"kind": "value", "type": "string", "value": "bar"}
					]
				},
				"result": {
					"type": "string",
					"value": "value"
				}
			}
		`, string(raw))
	})

	t.Run("In", func(t *testing.T) {
		r, err := New(
			In(VarStr("foo"), ValStr("bar"), ValStr("foo")),
			ReturnsStr("value"),
		)
		require.NoError(t, err)

		raw, err := json.Marshal(r)
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"root": {
					"kind": "in",
					"operands": [
						{"kind": "variable", "type": "string", "name": "foo"},
						{"kind": "value", "type": "string", "value": "bar"},
						{"kind": "value", "type": "string", "value": "foo"}
					]
				},
				"result": {
					"type": "string",
					"value": "value"
				}
			}
		`, string(raw))
	})
}
