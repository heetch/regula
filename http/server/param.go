package server

import (
	"strconv"

	rerrors "github.com/heetch/regula/errors"
)

// params represents the parameters computed from the query string.
// It implements the rule.Params interface.
type params map[string]string

// GetString extracts a string parameter which corresponds to the given key.
func (p params) GetString(key string) (string, error) {
	s, ok := p[key]
	if !ok {
		return "", rerrors.ErrParamNotFound
	}

	return s, nil
}

// GetBool extracts a bool parameter which corresponds to the given key.
func (p params) GetBool(key string) (bool, error) {
	v, ok := p[key]
	if !ok {
		return false, rerrors.ErrParamNotFound
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, rerrors.ErrParamTypeMismatch
	}

	return b, nil
}

// GetInt64 extracts an int64 parameter which corresponds to the given key.
func (p params) GetInt64(key string) (int64, error) {
	v, ok := p[key]
	if !ok {
		return 0, rerrors.ErrParamNotFound
	}

	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, rerrors.ErrParamTypeMismatch
	}

	return i, nil
}

// GetFloat64 extracts a float64 parameter which corresponds to the given key.
func (p params) GetFloat64(key string) (float64, error) {
	v, ok := p[key]
	if !ok {
		return 0, rerrors.ErrParamNotFound
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, rerrors.ErrParamTypeMismatch
	}

	return f, err
}

// Keys returns the list of all the keys.
func (p params) Keys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}

	return keys
}

// EncodeValue returns the string representation of a value.
func (p params) EncodeValue(key string) (string, error) {
	v, ok := p[key]
	if !ok {
		return "", rerrors.ErrParamNotFound
	}

	return v, nil
}