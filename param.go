package regula

import (
	"strconv"

	"github.com/heetch/regula/errortype"
	"github.com/pkg/errors"
)

// Params is a map based param.Params implementation.
type Params map[string]interface{}

// GetString extracts a string parameter corresponding to the given key.
func (p Params) GetString(key string) (string, error) {
	v, ok := p[key]
	if !ok {
		return "", errortype.ErrParamNotFound
	}

	s, ok := v.(string)
	if !ok {
		return "", errortype.ErrParamTypeMismatch
	}

	return s, nil
}

// GetBool extracts a bool parameter corresponding to the given key.
func (p Params) GetBool(key string) (bool, error) {
	v, ok := p[key]
	if !ok {
		return false, errortype.ErrParamNotFound
	}

	b, ok := v.(bool)
	if !ok {
		return false, errortype.ErrParamTypeMismatch
	}

	return b, nil
}

// GetInt64 extracts an int64 parameter corresponding to the given key.
func (p Params) GetInt64(key string) (int64, error) {
	v, ok := p[key]
	if !ok {
		return 0, errortype.ErrParamNotFound
	}

	i, ok := v.(int64)
	if !ok {
		return 0, errortype.ErrParamTypeMismatch
	}

	return i, nil
}

// GetFloat64 extracts a float64 parameter corresponding to the given key.
func (p Params) GetFloat64(key string) (float64, error) {
	v, ok := p[key]
	if !ok {
		return 0, errortype.ErrParamNotFound
	}

	f, ok := v.(float64)
	if !ok {
		return 0, errortype.ErrParamTypeMismatch
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
		return "", errortype.ErrParamNotFound
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

func (p Params) AddParam(key string, value interface{}) (Params, error) {
	if _, exists := p[key]; exists {
		return nil, errors.Errorf("cannot create parameter %q as a parameter with that name already exists")
	}
	newParams := make(Params)
	newParams[key] = value
	for k, v := range p {
		newParams[k] = v
	}
	return newParams, nil
}
