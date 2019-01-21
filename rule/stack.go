package rule

import (
	"strconv"

	rerrors "github.com/heetch/regula/errors"
	"github.com/pkg/errors"
)

type stack struct {
	Params

	key   string
	value interface{}
}

func newStack(key string, value interface{}, p Params) *stack {
	return &stack{
		Params: p,
		key:    key,
		value:  value,
	}
}

// GetString extracts a string parameter corresponding to the given key.
func (s stack) GetString(key string) (string, error) {
	if key == s.key {
		s, ok := s.value.(string)
		if !ok {
			return "", rerrors.ErrParamTypeMismatch
		}

		return s, nil
	}
	return s.Params.GetString(key)
}

// GetBool extracts a bool parameter corresponding to the given key.
func (s stack) GetBool(key string) (bool, error) {
	if key == s.key {
		b, ok := s.value.(bool)
		if !ok {
			return false, rerrors.ErrParamTypeMismatch
		}
		return b, nil
	}
	return s.Params.GetBool(key)
}

// GetInt64 extracts an int64 parameter corresponding to the given key.
func (s stack) GetInt64(key string) (int64, error) {
	if key == s.key {
		i, ok := s.value.(int64)
		if !ok {
			return 0, rerrors.ErrParamTypeMismatch
		}
		return i, nil
	}
	return s.Params.GetInt64(key)
}

// GetFloat64 extracts a float64 parameter corresponding to the given key.
func (s stack) GetFloat64(key string) (float64, error) {
	if key == s.key {
		f, ok := s.value.(float64)
		if !ok {
			return 0, rerrors.ErrParamTypeMismatch
		}
		return f, nil
	}
	return s.Params.GetFloat64(key)
}

// Keys returns the list of all the keys.
func (s stack) Keys() []string {
	keys := s.Params.Keys()
	keys = append(keys, s.key)
	return keys
}

// EncodeValue returns the string representation of the selected value.
func (s stack) EncodeValue(key string) (string, error) {
	if key == s.key {
		switch t := s.value.(type) {
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
	return s.Params.EncodeValue(key)
}
