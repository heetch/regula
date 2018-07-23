package regula

import (
	"strconv"

	"github.com/pkg/errors"
)

// A ParamGetter is a set of parameters passed on rule evaluation.
// It provides type safe methods to query params.
type ParamGetter interface {
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
	Keys() []string
	EncodeValue(key string) (string, error)
}

// Params is a map based ParamGetter implementation.
type Params map[string]interface{}

// GetString extracts a string parameter corresponding to the given key.
func (p Params) GetString(key string) (string, error) {
	v, ok := p[key]
	if !ok {
		return "", ErrParamNotFound
	}

	s, ok := v.(string)
	if !ok {
		return "", ErrParamTypeMismatch
	}

	return s, nil
}

// GetBool extracts a bool parameter corresponding to the given key.
func (p Params) GetBool(key string) (bool, error) {
	v, ok := p[key]
	if !ok {
		return false, ErrParamNotFound
	}

	b, ok := v.(bool)
	if !ok {
		return false, ErrParamTypeMismatch
	}

	return b, nil
}

// GetInt64 extracts an int64 parameter corresponding to the given key.
func (p Params) GetInt64(key string) (int64, error) {
	v, ok := p[key]
	if !ok {
		return 0, ErrParamNotFound
	}

	i, ok := v.(int64)
	if !ok {
		return 0, ErrParamTypeMismatch
	}

	return i, nil
}

// GetFloat64 extracts a float64 parameter corresponding to the given key.
func (p Params) GetFloat64(key string) (float64, error) {
	v, ok := p[key]
	if !ok {
		return 0, ErrParamNotFound
	}

	f, ok := v.(float64)
	if !ok {
		return 0, ErrParamTypeMismatch
	}

	return f, nil
}

// Keys returns the list of all the keys.
func (p Params) Keys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}

	return keys
}

// EncodeValue returns the string representation of the selected value.
func (p Params) EncodeValue(key string) (string, error) {
	v, ok := p[key]
	if !ok {
		return "", ErrParamNotFound
	}

	switch t := v.(type) {
	case string:
		return t, nil
	case int64:
		return strconv.FormatInt(t, 10), nil
	case float64:
		return strconv.FormatFloat(t, 'f', 6, 64), nil
	case bool:
		return strconv.FormatBool(t), nil
	default:
		return "", errors.Errorf("type %t is not supported", t)
	}
}
