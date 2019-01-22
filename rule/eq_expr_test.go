package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestEq(t *testing.T) {
	t.Run("Eval/Match", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
		m2 := mockExpr{val: rule.BoolValue(true)}
		params := regula.Params{"foo": "bar"}
		eq := rule.Eq(&m1, &m2)
		val, err := eq.Eval(params)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
		require.Equal(t, params, m1.lastParams)
		require.Equal(t, params, m2.lastParams)
	})

	t.Run("Eval/NoMatch", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
		m2 := mockExpr{val: rule.BoolValue(false)}
		eq := rule.Eq(&m1, &m2)
		val, err := eq.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(false), val)
	})
}

func TestLT(t *testing.T) {

	type testCase struct {
		Name     string
		Input    []rule.Expr
		Expected bool
	}

	type typeSuite struct {
		Name  string
		Cases []testCase
	}

	ts := []typeSuite{
		{
			Name: "Integer",
			Cases: []testCase{
				{
					Name:     "2 value, < ∴ True",
					Input:    []rule.Expr{rule.Int64Value(50), rule.Int64Value(60)},
					Expected: true,
				},
				{
					Name:     "3 value,< ∴ True",
					Input:    []rule.Expr{rule.Int64Value(50), rule.Int64Value(60), rule.Int64Value(70)},
					Expected: true,
				},
				{
					Name:     "2 value, = ∴ False",
					Input:    []rule.Expr{rule.Int64Value(50), rule.Int64Value(50)},
					Expected: false,
				},
				{
					Name:     "3 value, = ∴ False",
					Input:    []rule.Expr{rule.Int64Value(50), rule.Int64Value(50), rule.Int64Value(50)},
					Expected: false,
				},
				{
					Name:     "2 value, > ∴ False",
					Input:    []rule.Expr{rule.Int64Value(50), rule.Int64Value(50)},
					Expected: false,
				},
				{
					Name:     "3 value, > ∴ False",
					Input:    []rule.Expr{rule.Int64Value(50), rule.Int64Value(50), rule.Int64Value(50)},
					Expected: false,
				},
			},
		},
		{
			Name: "Float",
			Cases: []testCase{
				{
					Name:     "2 value, < ∴ True",
					Input:    []rule.Expr{rule.Float64Value(50.1), rule.Float64Value(50.2)},
					Expected: true,
				},
				{
					Name:     "3 value,< ∴ True",
					Input:    []rule.Expr{rule.Float64Value(50.1), rule.Float64Value(50.2), rule.Float64Value(50.3)},
					Expected: true,
				},
				{
					Name:     "2 value, = ∴ False",
					Input:    []rule.Expr{rule.Float64Value(50.1), rule.Float64Value(50.1)},
					Expected: false,
				},
				{
					Name:     "3 value, = ∴ False",
					Input:    []rule.Expr{rule.Float64Value(50.1), rule.Float64Value(50.1), rule.Float64Value(50.1)},
					Expected: false,
				},
				{
					Name:     "2 value, > ∴ False",
					Input:    []rule.Expr{rule.Float64Value(50.1), rule.Float64Value(50.1)},
					Expected: false,
				},
				{
					Name:     "3 value, > ∴ False",
					Input:    []rule.Expr{rule.Float64Value(50.1), rule.Float64Value(50.1), rule.Float64Value(50.1)},
					Expected: false,
				},
			},
		},
		{
			Name: "String",
			Cases: []testCase{
				{
					Name:     "Uppercase < Lowercase ∴ True",
					Input:    []rule.Expr{rule.StringValue("A"), rule.StringValue("a")},
					Expected: true,
				},
				{
					Name:     "ASCIIbetical ∴ True",
					Input:    []rule.Expr{rule.StringValue("0"), rule.StringValue("A")},
					Expected: true,
				},
				{
					Name:     "2 value, < ∴ True",
					Input:    []rule.Expr{rule.StringValue("ABBA"), rule.StringValue("Beegees")},
					Expected: true,
				},
				{
					Name:     "3 value,< ∴ True",
					Input:    []rule.Expr{rule.StringValue("ABBA"), rule.StringValue("Beegees"), rule.StringValue("Boney M")},
					Expected: true,
				},
				{
					Name:     "2 value, = ∴ False",
					Input:    []rule.Expr{rule.StringValue("ABBA"), rule.StringValue("ABBA")},
					Expected: false,
				},
				{
					Name:     "3 value, = ∴ False",
					Input:    []rule.Expr{rule.StringValue("ABBA"), rule.StringValue("ABBA"), rule.StringValue("ABBA")},
					Expected: false,
				},
				{
					Name:     "2 value, > ∴ False",
					Input:    []rule.Expr{rule.StringValue("ABBA"), rule.StringValue("ABBA")},
					Expected: false,
				},
				{
					Name:     "3 value, > ∴ False",
					Input:    []rule.Expr{rule.StringValue("ABBA"), rule.StringValue("ABBA"), rule.StringValue("ABBA")},
					Expected: false,
				},
			},
		},
		{
			Name: "Boolean",
			Cases: []testCase{
				{
					Name:     "True = True ∴ False",
					Input:    []rule.Expr{rule.BoolValue(true), rule.BoolValue(true)},
					Expected: false,
				},
				{
					Name:     "False = False ∴ False",
					Input:    []rule.Expr{rule.BoolValue(false), rule.BoolValue(false)},
					Expected: false,
				},
				{
					Name:     "False < True ∴ True",
					Input:    []rule.Expr{rule.BoolValue(false), rule.BoolValue(true)},
					Expected: true,
				},
				{
					Name:     "True > False ∴ False",
					Input:    []rule.Expr{rule.BoolValue(true), rule.BoolValue(false)},
					Expected: false,
				},
			},
		},
	}

	for _, s := range ts {
		t.Run(s.Name, func(t *testing.T) {
			for _, c := range s.Cases {
				t.Run(c.Name, func(t *testing.T) {
					lt := rule.LT(c.Input...)
					val, err := lt.Eval(nil)
					require.NoError(t, err)
					require.True(t, val.Same(rule.BoolValue(c.Expected)))
				})
			}
		})
	}
}
