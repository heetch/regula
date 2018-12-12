package sexpr_test

import (
	"bytes"
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/rule/sexpr"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Output rule.Expr
		Error  error
	}{
		{
			Name:   `simple bool`,
			Input:  `#true`,
			Output: rule.BoolValue(true),
			Error:  nil,
		},
		{
			Name:   `simple integer equals`,
			Input:  `(= 1 1)`,
			Output: rule.Eq(rule.Int64Value(1), rule.Int64Value(1)),
			Error:  nil,
		},
		{
			Name:   `simple float equals`,
			Input:  `(= 1.2 1.2)`,
			Output: rule.Eq(rule.Float64Value(1.2), rule.Float64Value(1.2)),
		},
		{
			Name:   `equals with number type promotion`,
			Input:  `(= 1 1.2)`,
			Output: rule.Eq(rule.IntToFloat(rule.Int64Value(1)), rule.Float64Value(1.2)),
		},
		{
			Name:   `simple parameter`,
			Input:  `foo`,
			Output: rule.BoolParam("foo"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			params := sexpr.Parameters{
				"foo": rule.BOOLEAN,
			}
			b := bytes.NewBufferString(c.Input)
			p := sexpr.NewParser(b)
			expr, err := p.Parse(params)
			switch c.Error {
			case nil:
				require.NoError(t, err)
				ce := expr.(rule.ComparableExpression)
				exp := c.Output.(rule.ComparableExpression)
				require.True(t, ce.Same(exp))
			default:
				require.EqualError(t, err, c.Error.Error())

			}
		})
	}

}
