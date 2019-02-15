package sexpr_test

import (
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/rule/sexpr"
	"github.com/stretchr/testify/require"
)

func TestPrettyPrintLet(t *testing.T) {
	e := rule.Let(rule.Int64Param("x"), rule.Int64Value(1), rule.Eq(rule.Int64Param("x"), rule.Int64Value(1)))
	result, err := sexpr.PrettyPrint(0, 79, e)
	require.NoError(t, err)
	require.Equal(t, "(let x 1\n     (= x 1))", result)
}

func TestPrettyPrintLetAsSubexpression(t *testing.T) {
	e := rule.Eq(rule.BoolValue(true), rule.Let(rule.Int64Param("x"), rule.Int64Value(1), rule.Eq(rule.Int64Param("x"), rule.Int64Value(1))))
	result, err := sexpr.PrettyPrint(0, 79, e)
	require.NoError(t, err)
	require.Equal(t, "(= #true (let x 1\n              (= x 1)))", result)
}

func TestPrettyPrintIf(t *testing.T) {
	e := rule.If(rule.Eq(rule.Int64Value(1), rule.Int64Value(2)), rule.StringValue("Yes"), rule.StringValue("No"))
	result, err := sexpr.PrettyPrint(0, 79, e)
	require.NoError(t, err)
	require.Equal(t, "(if (= 1 2)\n    \"Yes\"\n    \"No\")", result)
}

func TestPrettyPrintWithSingleLevelWrap(t *testing.T) {
	e := rule.Eq(rule.Int64Value(99), rule.Int64Value(99))
	result, err := sexpr.PrettyPrint(0, 7, e)
	require.NoError(t, err)
	require.Equal(t, "(= 99\n   99)", result)
}

func TestPrettyPrintWithWrapping(t *testing.T) {
	e := rule.Eq(rule.Div(rule.Int64Value(99), rule.Add(rule.Float64Value(23.22), rule.Int64Value(3))), rule.Float64Value(77.77))
	result, err := sexpr.PrettyPrint(0, 30, e)
	require.NoError(t, err)
	require.Equal(t, "(= (/ (int->float 99)\n      (+ 23.22 (int->float 3)))\n   77.77)", result)
}

func TestPrettyPrint(t *testing.T) {
	cases := []struct {
		name     string
		input    rule.Expr
		expected string
	}{
		{
			name:     "Int64Value",
			input:    rule.Int64Value(590),
			expected: "590",
		},
		{
			name:     "Negative Int64Value",
			input:    rule.Int64Value(-590),
			expected: "-590",
		},
		{
			name:     "Float64Value",
			input:    rule.Float64Value(72.8987),
			expected: "72.8987",
		},
		{
			name:     "Negative Float64Value",
			input:    rule.Float64Value(-72.8987),
			expected: "-72.8987",
		},
		{
			name:     "Boolean (true)",
			input:    rule.BoolValue(true),
			expected: "#true",
		},
		{
			name:     "Boolean (false)",
			input:    rule.BoolValue(false),
			expected: "#false",
		},
		{
			name:     "String",
			input:    rule.StringValue("foo"),
			expected: `"foo"`,
		},
		{
			// All parameters are just names when we print them.
			name:     "Parameter",
			input:    rule.StringParam("wibble"),
			expected: "wibble",
		},
		{
			name:     "Simple expression",
			input:    rule.Add(rule.Int64Value(10), rule.Int64Value(20)),
			expected: "(+ 10 20)",
		},
		{
			name:     "Complex expression",
			input:    rule.Eq(rule.Div(rule.Int64Value(99), rule.Add(rule.Float64Value(23.22), rule.Int64Value(3))), rule.Float64Value(77.77)),
			expected: `(= (/ (int->float 99) (+ 23.22 (int->float 3))) 77.77)`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := sexpr.PrettyPrint(0, 79, c.input)
			require.NoError(t, err)
			require.Equal(t, c.expected, result)
		})
	}
}
