package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestRuleEval(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		tests := []struct {
			expr   rule.Expr
			params rule.Params
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
			params rule.Params
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
