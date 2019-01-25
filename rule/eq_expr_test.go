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

type testCase struct {
	Name     string
	Input    []rule.Expr
	Expected bool
}

type typeSuite struct {
	Name  string
	Cases []testCase
}

// Test each case in the typeSuite.
func (ts typeSuite) Run(expr func(...rule.Expr) rule.Expr, t *testing.T) {
	t.Run(ts.Name, func(t *testing.T) {
		for _, c := range ts.Cases {
			t.Run(c.Name, func(t *testing.T) {
				ex := expr(c.Input...)
				val, err := ex.Eval(nil)
				require.NoError(t, err)
				require.True(t, val.Same(rule.BoolValue(c.Expected)))
			})
		}
	})
}

type comparitorTest struct {
	Expr   func(...rule.Expr) rule.Expr
	Suites []typeSuite
}

//
func (ct comparitorTest) Run(t *testing.T) {
	for _, s := range ct.Suites {
		s.Run(ct.Expr, t)
	}
}

func TestLT(t *testing.T) {
	ct := comparitorTest{
		Expr: rule.LT,
		Suites: []typeSuite{
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
		},
	}
	ct.Run(t)
}

func TestGT(t *testing.T) {
	ct := comparitorTest{
		Expr: rule.GT,
		Suites: []typeSuite{
			{
				Name: "Integer",
				Cases: []testCase{
					{
						Name:     "2 value, > ∴ True",
						Input:    []rule.Expr{rule.Int64Value(80), rule.Int64Value(60)},
						Expected: true,
					},
					{
						Name:     "3 value,> ∴ True",
						Input:    []rule.Expr{rule.Int64Value(80), rule.Int64Value(60), rule.Int64Value(40)},
						Expected: true,
					},
					{
						Name:     "2 value, = ∴ False",
						Input:    []rule.Expr{rule.Int64Value(80), rule.Int64Value(80)},
						Expected: false,
					},
					{
						Name:     "3 value, = ∴ False",
						Input:    []rule.Expr{rule.Int64Value(80), rule.Int64Value(80), rule.Int64Value(80)},
						Expected: false,
					},
					{
						Name:     "2 value, < ∴ False",
						Input:    []rule.Expr{rule.Int64Value(80), rule.Int64Value(80)},
						Expected: false,
					},
					{
						Name:     "3 value, < ∴ False",
						Input:    []rule.Expr{rule.Int64Value(80), rule.Int64Value(80), rule.Int64Value(80)},
						Expected: false,
					},
				},
			},
			{
				Name: "Float",
				Cases: []testCase{
					{
						Name:     "2 value, > ∴ True",
						Input:    []rule.Expr{rule.Float64Value(80.7), rule.Float64Value(80.2)},
						Expected: true,
					},
					{
						Name:     "3 value,> ∴ True",
						Input:    []rule.Expr{rule.Float64Value(80.7), rule.Float64Value(80.2), rule.Float64Value(80.1)},
						Expected: true,
					},
					{
						Name:     "2 value, = ∴ False",
						Input:    []rule.Expr{rule.Float64Value(80.7), rule.Float64Value(80.7)},
						Expected: false,
					},
					{
						Name:     "3 value, = ∴ False",
						Input:    []rule.Expr{rule.Float64Value(80.7), rule.Float64Value(80.7), rule.Float64Value(80.7)},
						Expected: false,
					},
					{
						Name:     "2 value, < ∴ False",
						Input:    []rule.Expr{rule.Float64Value(80.7), rule.Float64Value(80.7)},
						Expected: false,
					},
					{
						Name:     "3 value, < ∴ False",
						Input:    []rule.Expr{rule.Float64Value(80.7), rule.Float64Value(80.7), rule.Float64Value(80.7)},
						Expected: false,
					},
				},
			},
			{
				Name: "String",
				Cases: []testCase{
					{
						Name:     "Lowercase > Uppercase ∴ True",
						Input:    []rule.Expr{rule.StringValue("a"), rule.StringValue("A")},
						Expected: true,
					},
					{
						Name:     "ASCIIbetical ∴ True",
						Input:    []rule.Expr{rule.StringValue("A"), rule.StringValue("0")},
						Expected: true,
					},
					{
						Name:     "2 value, > ∴ True",
						Input:    []rule.Expr{rule.StringValue("ZZ Top"), rule.StringValue("T-Rex")},
						Expected: true,
					},
					{
						Name:     "3 value,> ∴ True",
						Input:    []rule.Expr{rule.StringValue("ZZ Top"), rule.StringValue("Uriah Heap"), rule.StringValue("T-Rex")},
						Expected: true,
					},
					{
						Name:     "2 value, = ∴ False",
						Input:    []rule.Expr{rule.StringValue("ZZ Top"), rule.StringValue("ZZ Top")},
						Expected: false,
					},
					{
						Name:     "3 value, = ∴ False",
						Input:    []rule.Expr{rule.StringValue("ZZ Top"), rule.StringValue("ZZ Top"), rule.StringValue("ZZ Top")},
						Expected: false,
					},
					{
						Name:     "2 value, < ∴ False",
						Input:    []rule.Expr{rule.StringValue("ZZ Top"), rule.StringValue("zz Top")},
						Expected: false,
					},
					{
						Name:     "3 value, < ∴ False",
						Input:    []rule.Expr{rule.StringValue("ZZ Top"), rule.StringValue("ZZ Top"), rule.StringValue("zz Top")},
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
						Name:     "False < True ∴ False",
						Input:    []rule.Expr{rule.BoolValue(false), rule.BoolValue(true)},
						Expected: false,
					},
					{
						Name:     "#True > False ∴ True",
						Input:    []rule.Expr{rule.BoolValue(true), rule.BoolValue(false)},
						Expected: true,
					},
				},
			},
		},
	}
	ct.Run(t)
}

func TestLTE(t *testing.T) {
	ct := comparitorTest{
		Expr: rule.LTE,
		Suites: []typeSuite{
			{
				Name: "Integer",
				Cases: []testCase{
					{
						Name:     "2 params, a < b ∴ true",
						Input:    []rule.Expr{rule.Int64Value(10), rule.Int64Value(20)},
						Expected: true,
					},
					{
						Name:     "3 params, a < b < c ∴ true",
						Input:    []rule.Expr{rule.Int64Value(10), rule.Int64Value(20), rule.Int64Value(30)},
						Expected: true,
					},
					{
						Name:     "2 params, a = b ∴ true",
						Input:    []rule.Expr{rule.Int64Value(10), rule.Int64Value(10)},
						Expected: true,
					},
					{
						Name:     "3 params, a = b = c ∴ true",
						Input:    []rule.Expr{rule.Int64Value(10), rule.Int64Value(10), rule.Int64Value(10)},
						Expected: true,
					},
					{
						Name:     "3 params, a < b = c ∴ true",
						Input:    []rule.Expr{rule.Int64Value(5), rule.Int64Value(10), rule.Int64Value(10)},
						Expected: true,
					},
					{
						Name:     "3 params, a = b < c ∴ true",
						Input:    []rule.Expr{rule.Int64Value(5), rule.Int64Value(5), rule.Int64Value(10)},
						Expected: true,
					},
					{
						Name:     "2 params, a > b ∴ false",
						Input:    []rule.Expr{rule.Int64Value(30), rule.Int64Value(20)},
						Expected: false,
					},
					{
						Name:     "3 params, a = b > c ∴ false",
						Input:    []rule.Expr{rule.Int64Value(10), rule.Int64Value(20), rule.Int64Value(10)},
						Expected: false,
					},
				},
			},
			{
				Name: "Float",
				Cases: []testCase{
					{
						Name:     "2 params, a < b ∴ true",
						Input:    []rule.Expr{rule.Float64Value(10.1), rule.Float64Value(10.2)},
						Expected: true,
					},
					{
						Name:     "3 params, a < b < c ∴ true",
						Input:    []rule.Expr{rule.Float64Value(10.1), rule.Float64Value(10.2), rule.Float64Value(10.3)},
						Expected: true,
					},
					{
						Name:     "2 params, a = b ∴ true",
						Input:    []rule.Expr{rule.Float64Value(10.1), rule.Float64Value(10.1)},
						Expected: true,
					},
					{
						Name:     "3 params, a = b = c ∴ true",
						Input:    []rule.Expr{rule.Float64Value(10.1), rule.Float64Value(10.1), rule.Float64Value(10.1)},
						Expected: true,
					},
					{
						Name:     "3 params, a < b = c ∴ true",
						Input:    []rule.Expr{rule.Float64Value(9.9), rule.Float64Value(10.1), rule.Float64Value(10.1)},
						Expected: true,
					},
					{
						Name:     "3 params, a = b < c ∴ true",
						Input:    []rule.Expr{rule.Float64Value(9.9), rule.Float64Value(9.9), rule.Float64Value(10.1)},
						Expected: true,
					},
					{
						Name:     "2 params, a > b ∴ false",
						Input:    []rule.Expr{rule.Float64Value(10.3), rule.Float64Value(10.2)},
						Expected: false,
					},
					{
						Name:     "3 params, a = b > c ∴ false",
						Input:    []rule.Expr{rule.Float64Value(10.1), rule.Float64Value(10.2), rule.Float64Value(10.1)},
						Expected: false,
					},
				},
			},
			{
				Name: "String",
				Cases: []testCase{
					{
						Name: "2 params, a < b ∴ true",
						Input: []rule.Expr{
							rule.StringValue("ABBA"),
							rule.StringValue("BeeGees"),
						},
						Expected: true,
					},
					{
						Name: "3 params, a < b < c∴ true",
						Input: []rule.Expr{
							rule.StringValue("ABBA"),
							rule.StringValue("BeeGees"),
							rule.StringValue("Chumbawumba"),
						},
						Expected: true,
					},
					{
						Name: "2 params, a = b ∴ true",
						Input: []rule.Expr{
							rule.StringValue("ABBA"),
							rule.StringValue("ABBA"),
						},
						Expected: true,
					},
					{
						Name: "3 params, a = b = c ∴ true",
						Input: []rule.Expr{
							rule.StringValue("ABBA"),
							rule.StringValue("ABBA"),
							rule.StringValue("ABBA"),
						},
						Expected: true,
					},
					{
						Name: "3 params, a < b = c ∴ true",
						Input: []rule.Expr{
							rule.StringValue("ABBA"),
							rule.StringValue("BeeGees"),
							rule.StringValue("BeeGees"),
						},
						Expected: true,
					},
					{
						Name: "3 params, a = b < c ∴ true",
						Input: []rule.Expr{
							rule.StringValue("ABBA"),
							rule.StringValue("ABBA"),
							rule.StringValue("BeeGees"),
						},
						Expected: true,
					},
					{
						Name: "2 params, a > b ∴ false",
						Input: []rule.Expr{
							rule.StringValue("BeeGees"),
							rule.StringValue("ABBA"),
						},
						Expected: false,
					},
					{
						Name: "3 params, a = b > c ∴ false",
						Input: []rule.Expr{
							rule.StringValue("BeeGees"),
							rule.StringValue("BeeGees"),
							rule.StringValue("ABBA"),
						},
						Expected: false,
					},
					{
						Name: "3 params, a > b = c ∴ false",
						Input: []rule.Expr{
							rule.StringValue("BeeGees"),
							rule.StringValue("ABBA"),
							rule.StringValue("ABBA"),
						},
						Expected: false,
					},
					{
						Name: "3 params, a > b > c ∴ false",
						Input: []rule.Expr{
							rule.StringValue("Chumbawumba"),
							rule.StringValue("BeeGees"),
							rule.StringValue("ABBA"),
						},
						Expected: false,
					},
				},
			},
		},
	}
	ct.Run(t)
}
