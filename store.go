package rules

// Store ...
type Store interface {
	Get(key string) (*RuleSet, error)
	FetchAll() error
}
