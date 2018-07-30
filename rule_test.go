package regula

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuleUnmarshalling(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var rule Rule

		err := rule.UnmarshalJSON([]byte(`{
			"result": {
				"data": "foo",
				"type": "string"
			},
			"root": {
				"kind": "eq",
				"operands": [
					{
						"kind": "value",
						"type": "string",
						"data": "bar"
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
								"data": "bar"
							}
						]
					}
				]
			}
		}`))
		require.NoError(t, err)
		require.Equal(t, "string", rule.Result.Type)
		require.Equal(t, "foo", rule.Result.Data)
		require.IsType(t, new(exprEq), rule.Root)
		eq := rule.Root.(*exprEq)
		require.Len(t, eq.Operands, 2)
		require.IsType(t, new(Value), eq.Operands[0])
		require.IsType(t, new(exprEq), eq.Operands[1])
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
						"data": "true"
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
			expr   Expr
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
			r := NewRule(test.expr, StringValue("matched"))
			res, err := r.Eval(test.params)
			require.NoError(t, err)
			require.Equal(t, "matched", res.Data)
			require.Equal(t, "string", res.Type)
		}
	})

	t.Run("Invalid return", func(t *testing.T) {
		tests := []struct {
			expr   Expr
			params Params
		}{
			{StringValue("foo"), nil},
			{StringParam("bar"), Params{"bar": "foo"}},
		}

		for _, test := range tests {
			r := NewRule(test.expr, StringValue("matched"))
			_, err := r.Eval(test.params)
			require.Error(t, err)
		}
	})
}

func TestRulesetEval(t *testing.T) {
	t.Run("Match string", func(t *testing.T) {
		r, err := NewStringRuleset(
			NewRule(Eq(StringValue("foo"), StringValue("bar")), StringValue("first")),
			NewRule(Eq(StringValue("foo"), StringValue("foo")), StringValue("second")),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "second", res.Data)
	})

	t.Run("Match bool", func(t *testing.T) {
		r, err := NewBoolRuleset(
			NewRule(Eq(StringValue("foo"), StringValue("bar")), BoolValue(false)),
			NewRule(Eq(StringValue("foo"), StringValue("foo")), BoolValue(true)),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "true", res.Data)
	})

	t.Run("Type mismatch", func(t *testing.T) {
		_, err := NewStringRuleset(
			NewRule(Eq(StringValue("foo"), StringValue("bar")), StringValue("first")),
			NewRule(Eq(StringValue("foo"), StringValue("foo")), BoolValue(true)),
		)
		require.Equal(t, ErrRulesetIncoherentType, err)
	})

	t.Run("No match", func(t *testing.T) {
		r, err := NewStringRuleset(
			NewRule(Eq(StringValue("foo"), StringValue("bar")), StringValue("first")),
			NewRule(Eq(StringValue("bar"), StringValue("foo")), StringValue("second")),
		)
		require.NoError(t, err)

		_, err = r.Eval(nil)
		require.Equal(t, ErrNoMatch, err)
	})

	t.Run("Default", func(t *testing.T) {
		r, err := NewStringRuleset(
			NewRule(Eq(StringValue("foo"), StringValue("bar")), StringValue("first")),
			NewRule(Eq(StringValue("bar"), StringValue("foo")), StringValue("second")),
			NewRule(True(), StringValue("default")),
		)
		require.NoError(t, err)

		res, err := r.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, "default", res.Data)
	})
}

func TestRulesetEncDec(t *testing.T) {
	r1, err := NewStringRuleset(
		NewRule(Eq(StringValue("foo"), StringValue("bar")), StringValue("first")),
		NewRule(Eq(StringValue("bar"), StringParam("foo")), StringValue("second")),
		NewRule(True(), StringValue("default")),
	)
	require.NoError(t, err)

	raw, err := json.Marshal(r1)
	require.NoError(t, err)

	var r2 Ruleset
	err = json.Unmarshal(raw, &r2)
	require.NoError(t, err)

	require.Equal(t, r1, &r2)
}
