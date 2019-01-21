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
	t.Run("Integer LessThan (True)", func(t *testing.T) {
		lt := rule.LT(rule.Int64Value(50), rule.Int64Value(60))
		val, err := lt.Eval(nil)
		require.NoError(t, err)
		require.True(t, val.Same(rule.BoolValue(true)))
	})
	t.Run("Integer LessThan Sequence(True)", func(t *testing.T) {
		lt := rule.LT(rule.Int64Value(50), rule.Int64Value(60), rule.Int64Value(61))
		val, err := lt.Eval(nil)
		require.NoError(t, err)
		require.True(t, val.Same(rule.BoolValue(true)))
	})

	t.Run("Integer LessThan (False)", func(t *testing.T) {
		lt := rule.LT(rule.Int64Value(70), rule.Int64Value(60))
		val, err := lt.Eval(nil)
		require.NoError(t, err)
		require.True(t, val.Same(rule.BoolValue(false)))
	})
	t.Run("Integer LessThan Sequence (False)", func(t *testing.T) {
		lt := rule.LT(rule.Int64Value(70), rule.Int64Value(60), rule.Int64Value(50))
		val, err := lt.Eval(nil)
		require.NoError(t, err)
		require.True(t, val.Same(rule.BoolValue(false)))
	})

}
