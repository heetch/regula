package rule_test

import (
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestNot(t *testing.T) {
	t.Run("Eval/true", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true), returnType: rule.BOOLEAN}
		not := rule.Not(&m1)
		val, err := not.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(false)}
		not := rule.Not(&m1)
		val, err := not.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockExpr{val: rule.StringValue("foobar")}
		not := rule.Not(&m1)
		_, err := not.Eval(nil)
		require.Error(t, err)
		require.Equal(t, 1, m1.evalCount)
	})
}

func TestAnd(t *testing.T) {
	t.Run("Eval/true", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
		m2 := mockExpr{val: rule.BoolValue(true)}
		and := rule.And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/short-circuit", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(false)}
		m2 := mockExpr{val: rule.BoolValue(true)}
		and := rule.And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 0, m2.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
		m2 := mockExpr{val: rule.BoolValue(false)}
		and := rule.And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockExpr{val: rule.StringValue("foobar")}
		m2 := mockExpr{val: rule.BoolValue(true)}
		and := rule.And(&m1, &m2)
		_, err := and.Eval(nil)
		require.Error(t, err)
	})
}

func TestOr(t *testing.T) {
	t.Run("Eval/true", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
		m2 := mockExpr{val: rule.BoolValue(true)}
		or := rule.Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 0, m2.evalCount)
	})

	t.Run("Eval/short-circuit", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(false)}
		m2 := mockExpr{val: rule.BoolValue(true)}
		or := rule.Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(false)}
		m2 := mockExpr{val: rule.BoolValue(false)}
		or := rule.Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockExpr{val: rule.StringValue("foobar")}
		m2 := mockExpr{val: rule.BoolValue(true)}
		or := rule.Or(&m1, &m2)
		_, err := or.Eval(nil)
		require.Error(t, err)
	})
}
