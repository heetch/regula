package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExprToInt64(t *testing.T) {
	cases := []struct {
		Name       string
		Expr       Expr
		Result     int64
		Error      string
		ParamCount int
	}{
		{
			Name:       "Int64Value",
			Expr:       Int64Value(10),
			Result:     10,
			Error:      "",
			ParamCount: 0,
		},
		{
			Name:       "Compound Expression",
			Expr:       Add(Int64Value(10), Int64Value(20)),
			Result:     30,
			Error:      "",
			ParamCount: 0,
		},
		{
			Name:       "With Param",
			Expr:       Int64Param("foo"),
			Result:     100,
			Error:      "",
			ParamCount: 1,
		},
		{
			Name:       "Not an Int",
			Expr:       StringValue("wibble"),
			Result:     -1,
			Error:      "strconv.ParseInt: parsing \"wibble\": invalid syntax",
			ParamCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			params := NewMockParams(
				map[string]interface{}{
					"foo": 100,
				})
			i, err := exprToInt64(tc.Expr, params)
			if len(tc.Error) != 0 {
				require.EqualError(t, err, tc.Error)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.Result, i)
			require.Equal(t, tc.ParamCount, params.IntCount)
		})
	}
}

func TestExprToFloat64(t *testing.T) {
	cases := []struct {
		Name       string
		Expr       Expr
		Result     float64
		Error      string
		ParamCount int
	}{
		{
			Name:       "Float64Value",
			Expr:       Float64Value(10.1),
			Result:     10.1,
			Error:      "",
			ParamCount: 0,
		},
		{
			Name:       "Compound Expression",
			Expr:       Add(Float64Value(10.1), Float64Value(20.2)),
			Result:     30.3,
			Error:      "",
			ParamCount: 0,
		},
		{
			Name:       "With Param",
			Expr:       Float64Param("foo"),
			Result:     10.1,
			Error:      "",
			ParamCount: 1,
		},
		{
			Name:       "Not a Float",
			Expr:       StringValue("boing"),
			Result:     -1,
			Error:      "strconv.ParseFloat: parsing \"boing\": invalid syntax",
			ParamCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			params := NewMockParams(
				map[string]interface{}{
					"foo": 10.1,
				})
			i, err := exprToFloat64(tc.Expr, params)
			if len(tc.Error) != 0 {
				require.EqualError(t, err, tc.Error)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.Result, i)
			require.Equal(t, tc.ParamCount, params.FloatCount)
		})
	}
}
