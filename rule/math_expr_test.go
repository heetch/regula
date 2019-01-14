package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	t.Run("Eval/Int64/OK", func(t *testing.T) {
		n1 := rule.Int64Value(1)
		n2 := rule.Int64Value(2)
		params := regula.Params{}
		add := rule.Add(n1, n2)
		val, err := add.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Int64Value(3)))
	})
	t.Run("Eval/Float64/OK", func(t *testing.T) {
		n1 := rule.Float64Value(1.1)
		n2 := rule.Float64Value(2.2)
		params := regula.Params{}
		add := rule.Add(n1, n2)
		val, err := add.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Float64Value(3.3)))
	})
}

func TestSub(t *testing.T) {
	t.Run("Eval/Int64/OK", func(t *testing.T) {
		n1 := rule.Int64Value(1)
		n2 := rule.Int64Value(1)
		params := regula.Params{}
		sub := rule.Sub(n1, n2)
		val, err := sub.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Int64Value(0)))
	})
	t.Run("Eval/Float64/OK", func(t *testing.T) {
		n1 := rule.Float64Value(2.2)
		n2 := rule.Float64Value(1.1)
		params := regula.Params{}
		sub := rule.Sub(n1, n2)
		val, err := sub.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Float64Value(1.1)))
	})

}

func TestMult(t *testing.T) {
	t.Run("Eval/Int64/OK", func(t *testing.T) {
		n1 := rule.Int64Value(11)
		n2 := rule.Int64Value(9)
		params := regula.Params{}
		mult := rule.Mult(n1, n2)
		val, err := mult.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Int64Value(99)))
	})
	t.Run("Eval/Float64/OK", func(t *testing.T) {
		n1 := rule.Float64Value(2.2)
		n2 := rule.Float64Value(1.1)
		params := regula.Params{}
		mult := rule.Mult(n1, n2)
		val, err := mult.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Float64Value(2.42)))
	})

}

func TestDiv(t *testing.T) {
	t.Run("Eval/Int64/OK", func(t *testing.T) {
		n1 := rule.Int64Value(10)
		n2 := rule.Int64Value(10)
		params := regula.Params{}
		div := rule.Div(n1, n2)
		val, err := div.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Int64Value(1)))
	})
	t.Run("Eval/Float64/OK", func(t *testing.T) {
		n1 := rule.Float64Value(2.2)
		n2 := rule.Float64Value(1.1)
		params := regula.Params{}
		div := rule.Div(n1, n2)
		val, err := div.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Float64Value(2.0)))
	})

}
