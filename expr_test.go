package regula

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockExpr struct {
	val        *Value
	err        error
	evalFn     func(params ParamGetter) (*Value, error)
	evalCount  int
	lastParams ParamGetter
}

func (m *mockExpr) Eval(params ParamGetter) (*Value, error) {
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
	t.Run("Marshalling", func(t *testing.T) {
		not := Not(new(mockExpr))

		raw, err := json.Marshal(not)
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"kind": "not",
				"operands": [
					{"kind": "mock"}
				]
			}
		`, string(raw))
	})

	t.Run("Unmarshalling", func(t *testing.T) {
		var not exprNot

		// enough operands
		err := not.UnmarshalJSON([]byte(`
			{
				"kind": "not",
				"operands": [
					{"kind": "value"}
				]
			}
		`))
		require.NoError(t, err)
		require.Len(t, not.Operands, 1)

		err = not.UnmarshalJSON([]byte(`
			{
				"kind": "not",
				"operands": [
					{"kind": "mock"}
				]
			}
		`))
		require.Error(t, err)
	})

	t.Run("Eval/true", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		not := Not(&m1)
		val, err := not.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(false)}
		not := Not(&m1)
		val, err := not.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockExpr{val: StringValue("foobar")}
		not := Not(&m1)
		_, err := not.Eval(nil)
		require.Error(t, err)
		require.Equal(t, 1, m1.evalCount)
	})
}

func TestAnd(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		and := And(new(mockExpr), new(mockExpr))

		raw, err := json.Marshal(and)
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"kind": "and",
				"operands": [
					{"kind": "mock"},
					{"kind": "mock"}
				]
			}
		`, string(raw))
	})

	t.Run("Unmarshalling", func(t *testing.T) {
		var and exprAnd

		// enough operands
		err := and.UnmarshalJSON([]byte(`
			{
				"kind": "and",
				"operands": [
					{"kind": "param"},
					{"kind": "value"}
				]
			}
		`))
		require.NoError(t, err)
		require.Len(t, and.Operands, 2)

		err = and.UnmarshalJSON([]byte(`
			{
				"kind": "and",
				"operands": [
					{"kind": "mock"}
				]
			}
		`))
		require.Error(t, err)
	})

	t.Run("Eval/true", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		m2 := mockExpr{val: BoolValue(true)}
		and := And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/short-circuit", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(false)}
		m2 := mockExpr{val: BoolValue(true)}
		and := And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 0, m2.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		m2 := mockExpr{val: BoolValue(false)}
		and := And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockExpr{val: StringValue("foobar")}
		m2 := mockExpr{val: BoolValue(true)}
		and := And(&m1, &m2)
		_, err := and.Eval(nil)
		require.Error(t, err)
	})
}

func TestOr(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		or := Or(new(mockExpr), new(mockExpr))

		raw, err := json.Marshal(or)
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"kind": "or",
				"operands": [
					{"kind": "mock"},
					{"kind": "mock"}
				]
			}
		`, string(raw))
	})

	t.Run("Unmarshalling", func(t *testing.T) {
		var or exprOr

		// enough operands
		err := or.UnmarshalJSON([]byte(`
			{
				"kind": "or",
				"operands": [
					{"kind": "param"},
					{"kind": "value"}
				]
			}
		`))
		require.NoError(t, err)
		require.Len(t, or.Operands, 2)

		err = or.UnmarshalJSON([]byte(`
			{
				"kind": "or",
				"operands": [
					{"kind": "mock"}
				]
			}
		`))
		require.Error(t, err)
	})

	t.Run("Eval/true", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		m2 := mockExpr{val: BoolValue(true)}
		or := Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 0, m2.evalCount)
	})

	t.Run("Eval/short-circuit", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(false)}
		m2 := mockExpr{val: BoolValue(true)}
		or := Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(false)}
		m2 := mockExpr{val: BoolValue(false)}
		or := Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockExpr{val: StringValue("foobar")}
		m2 := mockExpr{val: BoolValue(true)}
		or := Or(&m1, &m2)
		_, err := or.Eval(nil)
		require.Error(t, err)
	})
}

func TestEq(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		eq := Eq(new(mockExpr), new(mockExpr))

		raw, err := json.Marshal(eq)
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"kind": "eq",
				"operands": [
					{"kind": "mock"},
					{"kind": "mock"}
				]
			}
		`, string(raw))
	})

	t.Run("Unmarshalling", func(t *testing.T) {
		var eq exprEq

		// enough operands
		err := eq.UnmarshalJSON([]byte(`
			{
				"kind": "eq",
				"operands": [
					{"kind": "param"},
					{"kind": "value"}
				]
			}
		`))
		require.NoError(t, err)
		require.Len(t, eq.Operands, 2)

		err = eq.UnmarshalJSON([]byte(`
			{
				"kind": "eq",
				"operands": [
					{"kind": "mock"}
				]
			}
		`))
		require.Error(t, err)
	})

	t.Run("Eval/Match", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		m2 := mockExpr{val: BoolValue(true)}
		params := Params{"foo": "bar"}
		eq := Eq(&m1, &m2)
		val, err := eq.Eval(params)
		require.NoError(t, err)
		require.Equal(t, BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
		require.Equal(t, params, m1.lastParams)
		require.Equal(t, params, m2.lastParams)
	})

	t.Run("Eval/NoMatch", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		m2 := mockExpr{val: BoolValue(false)}
		eq := Eq(&m1, &m2)
		val, err := eq.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(false), val)
	})
}

func TestIn(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		eq := In(new(mockExpr), new(mockExpr))

		raw, err := json.Marshal(eq)
		require.NoError(t, err)
		require.JSONEq(t, `
			{
				"kind": "in",
				"operands": [
					{"kind": "mock"},
					{"kind": "mock"}
				]
			}
		`, string(raw))
	})

	t.Run("Unmarshalling", func(t *testing.T) {
		var in exprIn

		// enough operands
		err := in.UnmarshalJSON([]byte(`
			{
				"kind": "in",
				"operands": [
					{"kind": "param"},
					{"kind": "value"}
				]
			}
		`))
		require.NoError(t, err)
		require.Len(t, in.Operands, 2)

		err = in.UnmarshalJSON([]byte(`
			{
				"kind": "in",
				"operands": [
					{"kind": "mock"}
				]
			}
		`))
		require.Error(t, err)
	})

	t.Run("Eval/OK", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		m2 := mockExpr{val: BoolValue(true)}
		params := Params{"foo": "bar"}
		in := In(&m1, &m2)
		val, err := in.Eval(params)
		require.NoError(t, err)
		require.Equal(t, BoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
		require.Equal(t, params, m1.lastParams)
		require.Equal(t, params, m2.lastParams)
	})

	t.Run("Eval/Fail", func(t *testing.T) {
		m1 := mockExpr{val: BoolValue(true)}
		m2 := mockExpr{val: BoolValue(false)}
		eq := In(&m1, &m2)
		val, err := eq.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, BoolValue(false), val)
	})
}

func TestOperands(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var ops operands

		err := ops.UnmarshalJSON([]byte(`[]`))
		require.NoError(t, err)
		require.Empty(t, ops.Ops)
		require.Empty(t, ops.Exprs)
	})

	t.Run("Some Ops", func(t *testing.T) {
		var ops operands

		err := ops.UnmarshalJSON([]byte(`[
			{"kind": "value"},
			{"kind": "param"},
			{"kind": "eq","operands": [{"kind": "value"}, {"kind": "param"}]},
			{"kind": "in","operands": [{"kind": "value"}, {"kind": "param"}]}
		]`))
		require.NoError(t, err)
		require.Len(t, ops.Ops, 4)
		require.Len(t, ops.Exprs, 4)
	})
}

func TestParseExpr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		_, err := parseExpr("", []byte(``))
		require.Error(t, err)
	})

	t.Run("Unknown kind", func(t *testing.T) {
		_, err := parseExpr("kiwi", []byte(``))
		require.Error(t, err)
	})

	t.Run("OK", func(t *testing.T) {
		tests := []struct {
			kind string
			data []byte
			typ  interface{}
		}{
			{"eq", []byte(`{"kind": "eq","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprEq)},
			{"in", []byte(`{"kind":"in","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprIn)},
			{"not", []byte(`{"kind":"not","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprNot)},
			{"and", []byte(`{"kind":"and","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprAnd)},
			{"or", []byte(`{"kind":"or","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(exprOr)},
			{"param", []byte(`{"kind":"param"}`), new(exprParam)},
			{"value", []byte(`{"kind":"value"}`), new(Value)},
		}

		for _, test := range tests {
			n, err := parseExpr(test.kind, test.data)
			require.NoError(t, err)
			require.NotNil(t, n)
			require.IsType(t, test.typ, n)
		}
	})
}

func TestParam(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v := StringParam("foo")
		val, err := v.Eval(Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, StringValue("bar"), val)
	})

	t.Run("Not found", func(t *testing.T) {
		v := StringParam("foo")
		_, err := v.Eval(Params{
			"boo": "bar",
		})
		require.Error(t, err)
	})

	t.Run("Empty context", func(t *testing.T) {
		v := StringParam("foo")
		_, err := v.Eval(nil)
		require.Error(t, err)
	})
}

func TestValue(t *testing.T) {
	v1 := BoolValue(true)
	require.True(t, v1.Equal(v1))
	require.True(t, v1.Equal(BoolValue(true)))
	require.False(t, v1.Equal(BoolValue(false)))
	require.False(t, v1.Equal(StringValue("true")))
}
