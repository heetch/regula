package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockExpr struct {
	val        *rule.Value
	err        error
	evalFn     func(params rule.Params) (*rule.Value, error)
	evalCount  int
	lastParams rule.Params
	returnType rule.Type
}

func (m *mockExpr) Eval(params rule.Params) (*rule.Value, error) {
	m.evalCount++
	m.lastParams = params

	if m.evalFn != nil {
		return m.evalFn(params)
	}

	return m.val, m.err
}

func (m *mockExpr) Same(e rule.ComparableExpression) bool {
	return false
}

func (m *mockExpr) GetKind() string {
	return "mock"
}

func (m *mockExpr) MarshalJSON() ([]byte, error) {
	return []byte(`{"kind": "mock"}`), nil
}

func (m *mockExpr) PushExpr(e rule.Expr) error {
	return nil
}

func (m *mockExpr) Finalise() error {
	return nil
}

//
func (m *mockExpr) Contract() rule.Contract {
	return rule.Contract{
		ReturnType: m.returnType,
		Terms: []rule.Term{
			{Type: rule.ANY, Cardinality: rule.MANY},
		},
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

// A ComparableExpression can check its equivalence to another ComparableExpression
func TestExprSame(t *testing.T) {
	expr1 := rule.Eq(
		rule.Int64Value(1),
		rule.Int64Value(1),
	).(rule.ComparableExpression)
	expr2 := rule.Eq(
		rule.Int64Value(1),
		rule.Int64Value(1),
	).(rule.ComparableExpression)
	expr3 := rule.Eq(
		rule.Int64Value(1),
		rule.Int64Value(2),
	).(rule.ComparableExpression)
	expr4 := rule.And(
		rule.BoolValue(true),
		rule.BoolValue(true),
	).(rule.ComparableExpression)

	assert.True(t, expr1.Same(expr2))
	assert.False(t, expr1.Same(expr3))
	assert.False(t, expr1.Same(expr4))
}

// A ComparableExpression can check it equivalence to a complete AST.
func TestExprTreeSame(t *testing.T) {
	expr1 := rule.Eq(
		rule.And(
			rule.And(
				rule.BoolValue(true),
				rule.BoolValue(true),
			),
			rule.Or(
				rule.BoolValue(false),
				rule.BoolValue(true),
			),
			rule.BoolValue(true),
		),
		rule.BoolValue(false),
	).(rule.ComparableExpression)
	expr2 := rule.Eq(
		rule.And(
			rule.And(
				rule.BoolValue(true),
				// This is flipped to false
				rule.BoolValue(false),
			),
			rule.Or(
				rule.BoolValue(false),
				rule.BoolValue(true),
			),
			rule.BoolValue(true),
		),
		rule.BoolValue(false),
	).(rule.ComparableExpression)

	expr3 := rule.Eq(
		rule.And(
			rule.And(
				rule.BoolValue(true),
				rule.BoolValue(true),
			),
			rule.Or(
				rule.BoolValue(false),
				rule.BoolValue(true),
			),
			// This is flipped to false
			rule.BoolValue(false),
		),
		rule.BoolValue(false),
	).(rule.ComparableExpression)

	expr4 := rule.Not(rule.BoolValue(false)).(rule.ComparableExpression)

	assert.True(t, expr1.Same(expr1))
	assert.False(t, expr1.Same(expr2))
	assert.False(t, expr1.Same(expr3))
	assert.False(t, expr1.Same(expr4))
}

// A Value is a ComparableExpression
func TestValueSameness(t *testing.T) {
	v1 := rule.StringValue("foo")
	v2 := rule.StringValue("bar")
	v3 := rule.Int64Value(42)
	require.True(t, v1.Same(v1))
	require.False(t, v1.Same(v2))
	require.False(t, v1.Same(v3))
}

// A Param is a ComparableExpression
func TestParamSameness(t *testing.T) {
	p1 := rule.StringParam("bob")
	p2 := rule.StringParam("dave")
	p3 := rule.Float64Param("bob")
	require.True(t, p1.Same(p1))
	require.False(t, p1.Same(p2))
	require.False(t, p1.Same(p3))
}
