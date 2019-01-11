package rule_test

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	t.Run("Eval/Int64/OK", func(t *testing.T) {
		n1 := rule.Int64Value(1)
		n2 := rule.Int64Value(2)
		params := regula.Params{}
		add := rule.Add(n1, n2)
		val, err := add.Eval(params)
		require.NoError(t, err)
		require.True(t, val.Same(rule.Int64Value(3)))

	})

}
