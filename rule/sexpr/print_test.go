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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := sexpr.Print(c.input)
			require.Equal(t, c.expected, result)
		})
	}
}
