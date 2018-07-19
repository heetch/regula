package regula

// A ParamGetter enables to extract a parameter (key) of a specific type from a map.
type ParamGetter interface {
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
}

// Params is a set of variables passed on rule evaluation.
type Params map[string]interface{}

// GetString extracts a string parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrParamTypeMismatch.
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

// GetBool extracts a bool parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrParamTypeMismatch.
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

// GetInt64 extracts an int64 parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrParamTypeMismatch.
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

// GetFloat64 extracts a float64 parameter which corresponds to the given key.
// If the key doesn't exist, it returns ErrParamNotFound.
// If the type assertion fails, it returns ErrParamTypeMismatch.
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
