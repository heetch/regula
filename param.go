package regula

// A ParamGetter is a set of parameters passed on rule evaluation.
// It provides type safe methods to query params.
type ParamGetter interface {
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
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
