package param

// A Params is a set of parameters passed on rule evaluation.
// It provides type safe methods to query params.
type Params interface {
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
	Keys() []string
	EncodeValue(key string) (string, error)
}
