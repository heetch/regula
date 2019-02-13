package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestFNV(t *testing.T) {
	cases := []struct {
		name   string
		val    rule.Expr
		result int64
	}{
		{
			name:   "Int64Value",
			val:    rule.Int64Value(1234),
			result: 2179869525,
		},
		{
			name:   "Float64Value",
			val:    rule.Float64Value(1234.1234),
			result: 566939793,
		},
		{
			name:   "StringValue",
			val:    rule.StringValue("travelling in style"),
			result: 536463009,
		},
		{
			name:   "BoolValue (true)",
			val:    rule.BoolValue(true),
			result: 3053630529,
		},
		{
			name:   "BoolValue (false)",
			val:    rule.BoolValue(false),
			result: 2452206122,
		},
	}
	params := regula.Params{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hash := rule.FNV(tc.val)
			result, err := hash.Eval(params)
			require.NoError(t, err)
			require.Equal(t, rule.Int64Value(tc.result), result)
		})
	}
}
