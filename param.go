package regula

import (
	"strconv"

	rerrors "github.com/heetch/regula/errors"
	"github.com/pkg/errors"
)

// Params is a map based rule.Params implementation.
type Params map[string]interface{}

// GetString extracts a string parameter corresponding to the given key.
func (p Params) GetString(key string) (string, error) {
	v, ok := p[key]
	if !ok {
		return "", rerrors.ErrParamNotFound
	}

	s, ok := v.(string)
	if !ok {
		return "", rerrors.ErrParamTypeMismatch
	}

	return s, nil
}

// GetBool extracts a bool parameter corresponding to the given key.
func (p Params) GetBool(key string) (bool, error) {
	v, ok := p[key]
	if !ok {
		return false, rerrors.ErrParamNotFound
	}

	b, ok := v.(bool)
	if !ok {
		return false, rerrors.ErrParamTypeMismatch
	}

	return b, nil
}

// GetInt64 extracts an int64 parameter corresponding to the given key.
func (p Params) GetInt64(key string) (int64, error) {
	v, ok := p[key]
	if !ok {
		return 0, rerrors.ErrParamNotFound
	}

	i, ok := v.(int64)
	if !ok {
		return 0, rerrors.ErrParamTypeMismatch
	}

	return i, nil
}

// GetFloat64 extracts a float64 parameter corresponding to the given key.
func (p Params) GetFloat64(key string) (float64, error) {
	v, ok := p[key]
	if !ok {
		return 0, rerrors.ErrParamNotFound
	}

	f, ok := v.(float64)
	if !ok {
		return 0, rerrors.ErrParamTypeMismatch
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
		return "", rerrors.ErrParamNotFound
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
