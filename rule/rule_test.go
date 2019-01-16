package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/param"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestRuleEval(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		tests := []struct {
			expr   rule.Expr
			params param.Params
		}{
			{rule.Eq(rule.StringValue("foo"), rule.StringValue("foo")), nil},
			{rule.Eq(rule.StringValue("foo"), rule.StringParam("bar")), regula.Params{"bar": "foo"}},
			{rule.In(rule.StringValue("foo"), rule.StringParam("bar")), regula.Params{"bar": "foo"}},
			{
				rule.Eq(
					rule.Eq(rule.StringValue("bar"), rule.StringValue("bar")),
					rule.Eq(rule.StringValue("foo"), rule.StringValue("foo")),
				),
				nil,
			},
			{rule.True(), nil},
		}

		for _, test := range tests {
			r := rule.New(test.expr, rule.StringValue("matched"))
			res, err := r.Eval(test.params)
			require.NoError(t, err)
			require.Equal(t, "matched", res.Data)
			require.Equal(t, "string", res.Type)
		}
	})

	t.Run("Invalid return", func(t *testing.T) {
		tests := []struct {
			expr   rule.Expr
			params param.Params
		}{
			{rule.StringValue("foo"), nil},
			{rule.StringParam("bar"), regula.Params{"bar": "foo"}},
		}

		for _, test := range tests {
			r := rule.New(test.expr, rule.StringValue("matched"))
			_, err := r.Eval(test.params)
			require.Error(t, err)
		}
	})
}

func TestRuleParams(t *testing.T) {
	tc := []struct {
		rule   *rule.Rule
		params []rule.Param
	}{
		{rule.New(rule.True(), rule.StringValue("result")), nil},
		{
			rule.New(rule.StringParam("a"), rule.StringValue("result")),
			[]rule.Param{*rule.StringParam("a")},
		},
		{
			rule.New(
				rule.And(
					rule.Eq(rule.Int64Param("a"), rule.Int64Value(10)),
					rule.Eq(rule.BoolParam("b"), rule.BoolValue(true)),
				), rule.StringValue("result")),
			[]rule.Param{*rule.Int64Param("a"), *rule.BoolParam("b")},
		},
	}

	for _, tt := range tc {
		params := tt.rule.Params()
		require.Equal(t, tt.params, params)
	}
}
