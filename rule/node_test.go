package rule

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockNode struct {
	val        *Value
	err        error
	evalFn     func(params Params) (*Value, error)
	evalCount  int
	lastParams Params
}

func (m *mockNode) Eval(params Params) (*Value, error) {
	m.evalCount++
	m.lastParams = params

	if m.evalFn != nil {
		return m.evalFn(params)
	}

	return m.val, m.err
}

func (m *mockNode) MarshalJSON() ([]byte, error) {
	return []byte(`{"kind": "mock"}`), nil
}

func TestNot(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		not := Not(new(mockNode))

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
		var not nodeNot

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
		m1 := mockNode{val: NewBoolValue(true)}
		not := Not(&m1)
		val, err := not.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockNode{val: NewBoolValue(false)}
		not := Not(&m1)
		val, err := not.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockNode{val: NewStringValue("foobar")}
		not := Not(&m1)
		_, err := not.Eval(nil)
		require.Error(t, err)
		require.Equal(t, 1, m1.evalCount)
	})
}

func TestAnd(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		and := Or(new(mockNode), new(mockNode))

		raw, err := json.Marshal(and)
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
		var and nodeAnd

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
		m1 := mockNode{val: NewBoolValue(true)}
		m2 := mockNode{val: NewBoolValue(true)}
		and := And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/short-circuit", func(t *testing.T) {
		m1 := mockNode{val: NewBoolValue(false)}
		m2 := mockNode{val: NewBoolValue(true)}
		and := And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 0, m2.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockNode{val: NewBoolValue(true)}
		m2 := mockNode{val: NewBoolValue(false)}
		and := And(&m1, &m2)
		val, err := and.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockNode{val: NewStringValue("foobar")}
		m2 := mockNode{val: NewBoolValue(true)}
		and := And(&m1, &m2)
		_, err := and.Eval(nil)
		require.Error(t, err)
	})
}

func TestOr(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		or := Or(new(mockNode), new(mockNode))

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
		var or nodeOr

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
		m1 := mockNode{val: NewBoolValue(true)}
		m2 := mockNode{val: NewBoolValue(true)}
		or := Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 0, m2.evalCount)
	})

	t.Run("Eval/short-circuit", func(t *testing.T) {
		m1 := mockNode{val: NewBoolValue(false)}
		m2 := mockNode{val: NewBoolValue(true)}
		or := Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/false", func(t *testing.T) {
		m1 := mockNode{val: NewBoolValue(false)}
		m2 := mockNode{val: NewBoolValue(false)}
		or := Or(&m1, &m2)
		val, err := or.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(false), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
	})

	t.Run("Eval/error", func(t *testing.T) {
		m1 := mockNode{val: NewStringValue("foobar")}
		m2 := mockNode{val: NewBoolValue(true)}
		or := Or(&m1, &m2)
		_, err := or.Eval(nil)
		require.Error(t, err)
	})
}

func TestEq(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		eq := Eq(new(mockNode), new(mockNode))

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
		var eq nodeEq

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
		m1 := mockNode{val: NewBoolValue(true)}
		m2 := mockNode{val: NewBoolValue(true)}
		params := Params{"foo": "bar"}
		eq := Eq(&m1, &m2)
		val, err := eq.Eval(params)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
		require.Equal(t, params, m1.lastParams)
		require.Equal(t, params, m2.lastParams)
	})

	t.Run("Eval/NoMatch", func(t *testing.T) {
		m1 := mockNode{val: NewBoolValue(true)}
		m2 := mockNode{val: NewBoolValue(false)}
		eq := Eq(&m1, &m2)
		val, err := eq.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(false), val)
	})
}

func TestIn(t *testing.T) {
	t.Run("Marshalling", func(t *testing.T) {
		eq := In(new(mockNode), new(mockNode))

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
		var in nodeIn

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
		m1 := mockNode{val: NewBoolValue(true)}
		m2 := mockNode{val: NewBoolValue(true)}
		params := Params{"foo": "bar"}
		in := In(&m1, &m2)
		val, err := in.Eval(params)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(true), val)
		require.Equal(t, 1, m1.evalCount)
		require.Equal(t, 1, m2.evalCount)
		require.Equal(t, params, m1.lastParams)
		require.Equal(t, params, m2.lastParams)
	})

	t.Run("Eval/Fail", func(t *testing.T) {
		m1 := mockNode{val: NewBoolValue(true)}
		m2 := mockNode{val: NewBoolValue(false)}
		eq := In(&m1, &m2)
		val, err := eq.Eval(nil)
		require.NoError(t, err)
		require.Equal(t, NewBoolValue(false), val)
	})
}

func TestOperands(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var ops operands

		err := ops.UnmarshalJSON([]byte(`[]`))
		require.NoError(t, err)
		require.Empty(t, ops.Ops)
		require.Empty(t, ops.Nodes)
	})

	t.Run("Some Ops", func(t *testing.T) {
		var ops operands

		err := ops.UnmarshalJSON([]byte(`[
			{"kind": "value"},
			{"kind": "param"},
			{"kind": "true"},
			{"kind": "eq","operands": [{"kind": "value"}, {"kind": "param"}]},
			{"kind": "in","operands": [{"kind": "value"}, {"kind": "param"}]}
		]`))
		require.NoError(t, err)
		require.Len(t, ops.Ops, 5)
		require.Len(t, ops.Nodes, 5)
	})
}

func TestParseNode(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		_, err := parseNode("", []byte(``))
		require.Error(t, err)
	})

	t.Run("Unknown kind", func(t *testing.T) {
		_, err := parseNode("kiwi", []byte(``))
		require.Error(t, err)
	})

	t.Run("OK", func(t *testing.T) {
		tests := []struct {
			kind string
			data []byte
			typ  interface{}
		}{
			{"eq", []byte(`{"kind": "eq","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(nodeEq)},
			{"in", []byte(`{"kind":"in","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(nodeIn)},
			{"not", []byte(`{"kind":"not","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(nodeNot)},
			{"and", []byte(`{"kind":"and","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(nodeAnd)},
			{"or", []byte(`{"kind":"or","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(nodeOr)},
			{"param", []byte(`{"kind":"param"}`), new(nodeParam)},
			{"value", []byte(`{"kind":"value"}`), new(nodeValue)},
			{"true", []byte(`{"kind":"true"}`), new(nodeTrue)},
		}

		for _, test := range tests {
			n, err := parseNode(test.kind, test.data)
			require.NoError(t, err)
			require.NotNil(t, n)
			require.IsType(t, test.typ, n)
		}
	})
}

func TestVariable(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v := ParamStr("foo")
		val, err := v.Eval(Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, NewStringValue("bar"), val)
	})

	t.Run("Not found", func(t *testing.T) {
		v := ParamStr("foo")
		_, err := v.Eval(Params{
			"boo": "bar",
		})
		require.Error(t, err)
	})

	t.Run("Empty context", func(t *testing.T) {
		v := ParamStr("foo")
		_, err := v.Eval(nil)
		require.Error(t, err)
	})
}
