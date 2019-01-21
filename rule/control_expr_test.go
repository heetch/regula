package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestLet(t *testing.T) {
	params := make(regula.Params)
	l := rule.Let(rule.Int64Param("x"),
		rule.Add(rule.Int64Value(5),
			rule.Int64Value(5)),
		rule.Int64Param("x"))
	result, err := l.Eval(params)
	require.NoError(t, err)
	require.True(t, result.Same(rule.Int64Value(10)))
}

func TestIf(t *testing.T) {
	params := make(regula.Params)
	i := rule.If(rule.BoolValue(true),
		rule.BoolValue(true),
		rule.BoolValue(false))
	result, err := i.Eval(params)
	require.NoError(t, err)
	require.True(t, result.Same(rule.BoolValue(true)))

	i = rule.If(rule.BoolValue(false),
		rule.BoolValue(true),
		rule.BoolValue(false))
	result, err = i.Eval(params)
	require.NoError(t, err)
	require.True(t, result.Same(rule.BoolValue(false)))

}
