package rules

import (
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
			{"kind": "eq"},
			{"kind": "eq"},
			{"kind": "eq"}
		]`))
		require.NoError(t, err)
		require.Len(t, ops.Ops, 3)
		require.Len(t, ops.Nodes, 3)
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
		n, err := parseNode("eq", []byte(`{"kind":"eq"}`))
		require.NoError(t, err)
		require.NotNil(t, n)
		require.IsType(t, new(NodeEq), n)
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
				"kind": "eq"
			}
		}`))
		require.NoError(t, err)
		require.Equal(t, "foo", rule.Result.Value)
		require.IsType(t, new(NodeEq), rule.Root)
	})
}
