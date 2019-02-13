package rule_test

import (
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestPercentile(t *testing.T) {
	// "Bob Dylan" is in the 56th percentile, so this is true
	v1 := rule.StringValue("Bob Dylan")
	p := rule.Int64Value(96)
	perc := rule.Percentile(v1, p)
	res, err := perc.Eval(nil)
	require.NoError(t, err)
	require.Equal(t, rule.BoolValue(true), res)

	// "Joni Mitchell" is in the 97th percentile, so this is false
	v2 := rule.StringValue("Joni Mitchell")
	perc = rule.Percentile(v2, p)
	res, err = perc.Eval(nil)
	require.NoError(t, err)
	require.Equal(t, rule.BoolValue(false), res)
}
