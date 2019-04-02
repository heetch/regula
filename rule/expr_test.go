package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

type mockExpr struct {
	val        *rule.Value
	err        error
	evalFn     func(params rule.Params) (*rule.Value, error)
	evalCount  int
	lastParams rule.Params
}

func (m *mockExpr) Eval(params rule.Params) (*rule.Value, error) {
	m.evalCount++
	m.lastParams = params

	if m.evalFn != nil {
		return m.evalFn(params)
	}

	return m.val, m.err
}

func (m *mockExpr) MarshalJSON() ([]byte, error) {
	return []byte(`{"kind": "mock"}`), nil
}

func TestNot(t *testing.T) {
	t.Run("Eval/true", func(t *testing.T) {
		m1 := mockExpr{val: rule.BoolValue(true)}
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

func TestGt(t *testing.T) {
	cases := []struct {
		name     string
		m1       mockExpr
		m2       mockExpr
		expected *rule.Value
	}{
		{
			name:     "String: true",
			m1:       mockExpr{val: rule.StringValue("abd")},
			m2:       mockExpr{val: rule.StringValue("abc")},
			expected: rule.BoolValue(true),
		},
		{
			name:     "String: false",
			m1:       mockExpr{val: rule.StringValue("abc")},
			m2:       mockExpr{val: rule.StringValue("abd")},
			expected: rule.BoolValue(false),
		},
		{
			name:     "Bool: true",
			m1:       mockExpr{val: rule.BoolValue(true)},
			m2:       mockExpr{val: rule.BoolValue(false)},
			expected: rule.BoolValue(true),
		},
		{
			name:     "Bool: false#1",
			m1:       mockExpr{val: rule.BoolValue(true)},
			m2:       mockExpr{val: rule.BoolValue(true)},
			expected: rule.BoolValue(false),
		},
		{
			name:     "Bool: false#2",
			m1:       mockExpr{val: rule.BoolValue(false)},
			m2:       mockExpr{val: rule.BoolValue(true)},
			expected: rule.BoolValue(false),
		},
		{
			name:     "Int64: true",
			m1:       mockExpr{val: rule.Int64Value(12)},
			m2:       mockExpr{val: rule.Int64Value(11)},
			expected: rule.BoolValue(true),
		},
		{
			name:     "Int64: false",
			m1:       mockExpr{val: rule.Int64Value(12)},
			m2:       mockExpr{val: rule.Int64Value(12)},
			expected: rule.BoolValue(false),
		},
		{
			name:     "Float64: true",
			m1:       mockExpr{val: rule.Float64Value(12.1)},
			m2:       mockExpr{val: rule.Float64Value(12.0)},
			expected: rule.BoolValue(true),
		},
		{
			name:     "Float64: false",
			m1:       mockExpr{val: rule.Float64Value(12.0)},
			m2:       mockExpr{val: rule.Float64Value(12.1)},
			expected: rule.BoolValue(false),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gt := rule.Gt(&tc.m1, &tc.m2)
			val, err := gt.Eval(nil)
			require.NoError(t, err)
			require.Equal(t, tc.expected, val)
		})
	}
}

func TestParam(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v := rule.StringParam("foo")
		val, err := v.Eval(regula.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, rule.StringValue("bar"), val)
	})

	t.Run("Not found", func(t *testing.T) {
		v := rule.StringParam("foo")
		_, err := v.Eval(regula.Params{
			"boo": "bar",
		})
		require.Error(t, err)
	})

	t.Run("Empty context", func(t *testing.T) {
		v := rule.StringParam("foo")
		_, err := v.Eval(nil)
		require.Error(t, err)
	})
}

func TestValue(t *testing.T) {
	v1 := rule.BoolValue(true)
	require.True(t, v1.Equal(v1))
	require.True(t, v1.Equal(rule.BoolValue(true)))
	require.False(t, v1.Equal(rule.BoolValue(false)))
	require.False(t, v1.Equal(rule.StringValue("true")))
}
