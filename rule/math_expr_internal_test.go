package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mockParams struct {
	IntCount   int
	FloatCount int
}

// These methods are included here to make mockParams implment the

// rule.Params interface
func (p *mockParams) Keys() []string                                         { return nil }
func (p *mockParams) EncodeValue(key string) (string, error)                 { return "", nil }
func (p *mockParams) GetString(key string) (string, error)                   { return "", nil }
func (p *mockParams) GetBool(key string) (bool, error)                       { return true, nil }
func (p *mockParams) AddParam(key string, value interface{}) (Params, error) { return nil, nil }

func (p *mockParams) GetInt64(key string) (int64, error) {
	p.IntCount++
	return 100, nil
}
func (p *mockParams) GetFloat64(key string) (float64, error) {
	p.FloatCount++
	return 10.1, nil
}

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
			params := &mockParams{}
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
			params := &mockParams{}
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
