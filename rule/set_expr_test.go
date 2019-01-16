package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestIn(t *testing.T) {
	t.Run("Eval/OK", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
		m2 := mockExpr{val: rule.BoolValue(true)}
		params := regula.Params{"foo": "bar"}
		in := rule.In(&m1, &m2)
		val, err := in.Eval(params)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
		require.Equal(t, params, m1.lastParams)
		require.Equal(t, params, m2.lastParams)
	})

	t.Run("Eval/Fail", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
		m2 := mockExpr{val: rule.BoolValue(false)}
		eq := rule.In(&m1, &m2)
		val, err := eq.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, rule.BoolValue(false), val)
	})
}
