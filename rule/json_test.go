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
			{"kind": "eq","operands": [{"kind": "value"}, {"kind": "param"}]},
			{"kind": "in","operands": [{"kind": "value"}, {"kind": "param"}]}
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
			{"eq", []byte(`{"kind": "eq","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprEq)},
			{"in", []byte(`{"kind":"in","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprIn)},
			{"not", []byte(`{"kind":"not","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprNot)},
			{"and", []byte(`{"kind":"and","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprAnd)},
			{"or", []byte(`{"kind":"or","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprOr)},
			{"percentile", []byte(`{"kind":"percentile","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprPercentile)},
			{"gt", []byte(`{"kind":"gt","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprGT)},
			{"gte", []byte(`{"kind":"gte","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprGTE)},
			{"lt", []byte(`{"kind":"lt","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprLT)},
			{"lte", []byte(`{"kind":"lte","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprLTE)},
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
					Percentile(
						StringValue("Bob Dylan"),
						Int64Value(96),
					),
					Not(
						Eq(
							FNV(StringValue("Bob Dylan")),
							StringValue("Bob Dylan"),
						),
					),
					GT(
						Int64Value(11),
						Int64Value(10),
					),
					GTE(
						Int64Value(11),
						Int64Value(11),
					),
					LT(
						Int64Value(10),
						Int64Value(11),
					),
					LTE(
						Int64Value(10),
						Int64Value(10),
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
