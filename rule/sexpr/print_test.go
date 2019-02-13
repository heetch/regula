package sexpr_test

import (
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/rule/sexpr"
	"github.com/stretchr/testify/require"
)

func TestPrint(t *testing.T) {
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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := sexpr.Print(c.input)
			require.NoError(t, err)
			require.Equal(t, c.expected, result)
		})
	}
}
