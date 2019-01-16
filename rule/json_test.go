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
		require.Empty(t, ops.Exprs)
	})

	t.Run("Some Ops", func(t *testing.T) {
		var ops operands

		err := ops.UnmarshalJSON([]byte(`[
			{"kind": "value"},
			{"kind": "param"},
			{"kind": "eq","operands": [{"kind": "value", "type": "bool"}, {"kind": "param", "type": "bool"}]},
			{"kind": "in","operands": [{"kind": "value", "type": "string"}, {"kind": "param", "type": "string"}]}
		]`))
		require.NoError(t, err)
		require.Len(t, ops.Ops, 4)
		require.Len(t, ops.Exprs, 4)
	})
}

func TestUnmarshalExpr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		_, err := unmarshalExpr("", []byte(``))
		require.Error(t, err)
	})

	t.Run("Unknown kind", func(t *testing.T) {
		_, err := unmarshalExpr("kiwi", []byte(``))
		require.Error(t, err)
	})

	t.Run("OK", func(t *testing.T) {
		tests := []struct {
			kind string
			data []byte
			typ  interface{}
		}{
			{"eq", []byte(`{"kind": "eq","operands": [{"kind": "value", "type": "bool"}, {"kind": "param", "type": "bool"}]}`), new(exprEq)},
			{"in", []byte(`{"kind":"in","operands": [{"kind": "value", "type": "string"}, {"kind": "param", "type": "string"}]}`), new(exprIn)},
			{"not", []byte(`{"kind":"not","operands": [{"kind": "value", "type": "bool"}]}`), new(exprNot)},
			{"and", []byte(`{"kind":"and","operands": [{"kind": "value", "type": "bool"}, {"kind": "param", "type": "bool"}]}`), new(exprAnd)},
			{"or", []byte(`{"kind":"or","operands": [{"kind": "value", "type": "bool"}, {"kind": "param", "type": "bool"}]}`), new(exprOr)},
			{"param", []byte(`{"kind":"param"}`), new(Param)},
			{"value", []byte(`{"kind":"value"}`), new(Value)},
		}

		for _, test := range tests {
			n, err := unmarshalExpr(test.kind, test.data)
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
				"data": "foo",
				"type": "string"
			},
			"expr": {
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
		require.IsType(t, new(exprEq), rule.Expr)
		eq := rule.Expr.(*exprEq)
		require.Len(t, eq.operands, 2)
		require.IsType(t, new(Value), eq.operands[0])
		require.IsType(t, new(exprEq), eq.operands[1])
	})

	t.Run("EncDec", func(t *testing.T) {
		r1 := New(
			And(
				Or(
					Eq(
						StringValue("foo"),
						Int64Value(10),
						Float64Value(10),
						BoolValue(true),
					),
					In(
						StringParam("foo"),
						Int64Param("foo"),
						Float64Param("foo"),
						BoolParam("foo"),
					),
					Not(
						BoolValue(true),
					),
				),
				True(),
			),
			StringValue("ok"),
		)

		raw, err := json.Marshal(r1)
		require.NoError(t, err)

		var r2 Rule
		err = json.Unmarshal(raw, &r2)
		require.NoError(t, err)

		require.Equal(t, r1, &r2)
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

// An operator is a ComparableExpression
func TestOperatorSameness(t *testing.T) {
	o1 := &operator{
		operands: []Expr{BoolValue(true)},
		contract: Contract{OpCode: "not"}}
	o2 := Not(BoolValue(true)).(ComparableExpression)
	o3 := Or(BoolValue(true), BoolValue(false)).(ComparableExpression)
	require.True(t, o1.Same(o2))
	require.False(t, o1.Same(o3))
}

func TestOperatorPushExpr(t *testing.T) {
	not := newExprNot()
	not.PushExpr(BoolValue(false))

	expected := Not(BoolValue(false))
	notCE := ComparableExpression(not)
	expectedCE := expected.(ComparableExpression)

	require.True(t, notCE.Same(expectedCE))

}
