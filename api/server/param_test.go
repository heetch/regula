package server

import (
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestGetString(t *testing.T) {
	p := params{
		"string": "string",
		"bool":   "true",
	}

	t.Run("GetString - OK", func(t *testing.T) {
		v, err := p.GetString("string")
		require.NoError(t, err)
		require.Equal(t, "string", v)
	})

	t.Run("GetString - NOK - ErrParamNotFound", func(t *testing.T) {
		_, err := p.GetString("badkey")
		require.Error(t, err)
		require.EqualError(t, err, rule.ErrParamNotFound.Error())
	})
}

func TestGetBool(t *testing.T) {
	p := params{
		"bool":   "true",
		"string": "foo",
	}

	t.Run("GetBool - OK", func(t *testing.T) {
		v, err := p.GetBool("bool")
		require.NoError(t, err)
		require.Equal(t, true, v)
	})

	t.Run("GetBool - NOK - ErrParamNotFound", func(t *testing.T) {
		_, err := p.GetBool("badkey")
		require.Error(t, err)
		require.EqualError(t, err, rule.ErrParamNotFound.Error())
	})

	t.Run("GetBool - NOK - ErrTypeParamMismatch", func(t *testing.T) {
		_, err := p.GetBool("string")
		require.Error(t, err)
		require.EqualError(t, err, rule.ErrTypeParamMismatch.Error())
	})
}

func TestGetInt64(t *testing.T) {
	p := params{
		"int64":  "42",
		"string": "foo",
	}

	t.Run("GetInt64 - OK", func(t *testing.T) {
		v, err := p.GetInt64("int64")
		require.NoError(t, err)
		require.Equal(t, int64(42), v)
	})

	t.Run("GetInt64 - NOK - ErrParamNotFound", func(t *testing.T) {
		_, err := p.GetInt64("badkey")
		require.Error(t, err)
		require.EqualError(t, err, rule.ErrParamNotFound.Error())
	})

	t.Run("GetInt64 - NOK - ErrTypeParamMismatch", func(t *testing.T) {
		_, err := p.GetInt64("string")
		require.Error(t, err)
		require.EqualError(t, err, rule.ErrTypeParamMismatch.Error())
	})
}

func TestGetFloat64(t *testing.T) {
	p := params{
		"float64": "42.42",
		"string":  "foo",
	}

	t.Run("GetFloat64 - OK", func(t *testing.T) {
		v, err := p.GetFloat64("float64")
		require.NoError(t, err)
		require.Equal(t, 42.42, v)
	})

	t.Run("GetFloat64 - NOK - ErrParamNotFound", func(t *testing.T) {
		_, err := p.GetFloat64("badkey")
		require.Error(t, err)
		require.EqualError(t, err, rule.ErrParamNotFound.Error())
	})

	t.Run("GetFloat64 - NOK - ErrTypeParamMismatch", func(t *testing.T) {
		_, err := p.GetFloat64("string")
		require.Error(t, err)
		require.EqualError(t, err, rule.ErrTypeParamMismatch.Error())
	})
}
