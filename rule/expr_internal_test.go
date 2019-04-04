package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseBoolValues(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v1 := newValue("bool", "true")
		v2 := newValue("bool", "false")

		b1, b2, err := parseBoolValues(v1, v2)
		require.NoError(t, err)
		require.True(t, b1)
		require.False(t, b2)
	})

	t.Run("Fail 1st Value", func(t *testing.T) {
		v1 := newValue("bool", "foo")
		v2 := newValue("bool", "false")
		_, _, err := parseBoolValues(v1, v2)
		require.Error(t, err)
		require.Equal(t, err.Error(), `strconv.ParseBool: parsing "foo": invalid syntax`)
	})

	t.Run("Fail 2nd Value", func(t *testing.T) {
		v1 := newValue("bool", "true")
		v2 := newValue("bool", "bar")
		_, _, err := parseBoolValues(v1, v2)
		require.Error(t, err)
		require.Equal(t, err.Error(), `strconv.ParseBool: parsing "bar": invalid syntax`)
	})
}

func TestParseInt64Values(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v1 := newValue("int64", "123")
		v2 := newValue("int64", "456")

		i1, i2, err := parseInt64Values(v1, v2)
		require.NoError(t, err)
		require.Equal(t, int64(123), i1)
		require.Equal(t, int64(456), i2)
	})

	t.Run("Fail 1st Value", func(t *testing.T) {
		v1 := newValue("int64", "foo")
		v2 := newValue("int64", "456")

		_, _, err := parseInt64Values(v1, v2)
		require.Error(t, err)
		require.Equal(t, err.Error(), `strconv.ParseInt: parsing "foo": invalid syntax`)
	})

	t.Run("Fail 2nd Value", func(t *testing.T) {
		v1 := newValue("int64", "123")
		v2 := newValue("int64", "bar")

		_, _, err := parseInt64Values(v1, v2)
		require.Error(t, err)
		require.Equal(t, err.Error(), `strconv.ParseInt: parsing "bar": invalid syntax`)
	})
}

func TestParseFloat64Values(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v1 := newValue("float64", "12.3")
		v2 := newValue("float64", "45.6")

		f1, f2, err := parseFloat64Values(v1, v2)
		require.NoError(t, err)
		require.Equal(t, float64(12.3), f1)
		require.Equal(t, float64(45.6), f2)
	})

	t.Run("Fail 1st Value", func(t *testing.T) {
		v1 := newValue("float64", "foo")
		v2 := newValue("float64", "45.6")

		_, _, err := parseFloat64Values(v1, v2)
		require.Error(t, err)
		require.Equal(t, err.Error(), `strconv.ParseFloat: parsing "foo": invalid syntax`)
	})

	t.Run("Fail 2nd Value", func(t *testing.T) {
		v1 := newValue("float64", "12.3")
		v2 := newValue("float64", "bar")

		_, _, err := parseFloat64Values(v1, v2)
		require.Error(t, err)
		require.Equal(t, err.Error(), `strconv.ParseFloat: parsing "bar": invalid syntax`)
	})
}
