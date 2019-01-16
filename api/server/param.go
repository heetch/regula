package server

import (
	"strconv"

	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
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

// AddParam generates a new scoped Params which is a copy of the
// current Params with one additional Param mapping key to value.
// This is used by the let operator to create new lexical scopes.
func (p params) AddParam(key string, value interface{}) (rule.Params, error) {
	if _, exists := p[key]; exists {
		return nil, errors.Errorf("cannot create parameter %q as a parameter with that name already exists", key)
	}
	newParams := make(params)
	var newValue string
	switch t := value.(type) {
	case string:
		newValue = t
	case int64:
		newValue = strconv.FormatInt(t, 10)
	case float64:
		newValue = strconv.FormatFloat(t, 'f', 6, 64)
	case bool:
		newValue = strconv.FormatBool(t)
	default:
		return nil, errors.Errorf("type %t is not supported", t)
	}
	newParams[key] = newValue
	for k, v := range p {
		newParams[k] = v
	}
	return newParams, nil
}
