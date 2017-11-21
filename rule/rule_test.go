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
			{"kind": "true"}
		]`))
		require.NoError(t, err)
		require.Len(t, ops.Ops, 3)
		require.Len(t, ops.Nodes, 3)
	})
}

func TestParseOperator(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		_, err := parseOperator("", []byte(``))
		require.Error(t, err)
	})

	t.Run("Unknown kind", func(t *testing.T) {
		_, err := parseOperator("kiwi", []byte(``))
		require.Error(t, err)
	})

	t.Run("OK", func(t *testing.T) {
		tests := []struct {
			kind string
			data []byte
			typ  interface{}
		}{
			{"eq", []byte(`{"kind":"eq"}`), new(OpEq)},
			{"in", []byte(`{"kind":"in"}`), new(OpIn)},
		}

		for _, test := range tests {
			n, err := parseOperator(test.kind, test.data)
			require.NoError(t, err)
			require.NotNil(t, n)
			require.IsType(t, test.typ, n)
		}
	})
}

func TestParseOperand(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		_, err := parseOperand("", []byte(``))
		require.Error(t, err)
	})

	t.Run("Unknown kind", func(t *testing.T) {
		_, err := parseOperand("kiwi", []byte(``))
		require.Error(t, err)
	})

	t.Run("OK", func(t *testing.T) {
		tests := []struct {
			kind string
			data []byte
			typ  interface{}
		}{
			{"variable", []byte(`{"kind":"variable"}`), new(OpVariable)},
			{"value", []byte(`{"kind":"value"}`), new(OpValue)},
			{"true", []byte(`{"kind":"true"}`), new(OpTrue)},
		}

		for _, test := range tests {
			n, err := parseOperand(test.kind, test.data)
			require.NoError(t, err)
			require.NotNil(t, n)
			require.IsType(t, test.typ, n)
		}
	})
}

func TestRule(t *testing.T) {
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
					}
				]
			}
		}`))
		require.NoError(t, err)
		require.Equal(t, "foo", rule.Result.Value)
		require.IsType(t, new(OpEq), rule.Root)
		eq := rule.Root.(*OpEq)
		require.Len(t, eq.Operands, 2)
		require.IsType(t, new(OpVariable), eq.Operands[0])
		require.IsType(t, new(OpValue), eq.Operands[1])
	})
}

func TestRuleMarshalling(t *testing.T) {
	t.Run("Eq", func(t *testing.T) {
		r, err := New(
			Eq(Variable("foo", "string"), Value("bar", "string"), True()),
			Returns("value", "string"),
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
						{"kind": "value", "type": "string", "value": "bar"},
						{"kind": "true"}
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
			In(Variable("foo", "string"), Value("bar", "string"), Value("foo", "string")),
			Returns("value", "string"),
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
