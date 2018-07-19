package server

import (
	"strconv"

	"github.com/heetch/regula"
)

// params represents the parameters computed from the query string.
// It implements the regula.ParamGetter interface.
type params map[string]string

// GetString extracts a string parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrTypeParamMismatch.
func (p params) GetString(key string) (string, error) {
	s, ok := p[key]
	if !ok {
		return "", regula.ErrParamNotFound
	}

	return s, nil
}

// GetBool extracts a bool parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrTypeParamMismatch.
func (p params) GetBool(key string) (bool, error) {
	v, ok := p[key]
	if !ok {
		return false, regula.ErrParamNotFound
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, regula.ErrTypeParamMismatch
	}

	return b, nil
}

// GetInt64 extracts an int64 parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrTypeParamMismatch.
func (p params) GetInt64(key string) (int64, error) {
	v, ok := p[key]
	if !ok {
		return 0, regula.ErrParamNotFound
	}

	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, regula.ErrTypeParamMismatch
	}

	return i, nil
}

// GetFloat64 extracts a float64 parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrTypeParamMismatch.
func (p params) GetFloat64(key string) (float64, error) {
	v, ok := p[key]
	if !ok {
		return 0, regula.ErrParamNotFound
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, regula.ErrTypeParamMismatch
	}

	return f, err
}
