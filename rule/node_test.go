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
		var eq NodeEq

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

	t.Run("Eval/NotMatch", func(t *testing.T) {
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
		var in NodeIn

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
			{"eq", []byte(`{"kind": "eq","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(NodeEq)},
			{"in", []byte(`{"kind":"in","operands": [{"kind": "value"}, {"kind": "param"}]}`), new(NodeIn)},
			{"param", []byte(`{"kind":"param"}`), new(NodeParam)},
			{"value", []byte(`{"kind":"value"}`), new(NodeValue)},
			{"true", []byte(`{"kind":"true"}`), new(NodeTrue)},
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
