package rule_test

import (
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestIntToFloat(t *testing.T) {
	cast := rule.IntToFloat(rule.Int64Value(1))
	f, err := cast.Eval(nil)
	require.NoError(t, err)
	require.Equal(t, rule.Float64Value(1.0), f)
}
