package rules

import (
	"context"
	"strconv"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

var (
	// ErrRulesetNotFound must be returned when no ruleset is found for a given key.
	ErrRulesetNotFound = errors.New("ruleset not found")
	// ErrTypeMismatch is returned when the evaluated rule doesn't return the expected result type.
	ErrTypeMismatch = errors.New("type returned by rule doesn't match")
)

// Engine fetches the rules from the store and executes the selected ruleset.
type Engine struct {
	getter Getter
}

// NewEngine creates an Engine using the given getter.
func NewEngine(getter Getter) *Engine {
	return &Engine{
		getter: getter,
	}
}

// Get evaluates the ruleset associated with key and returns the result.
func (e *Engine) get(ctx context.Context, typ, key string, params rule.Params) (*rule.Value, error) {
	ruleset, err := e.getter.Get(ctx, key)
	if err != nil {
		if err == ErrRulesetNotFound {
			return nil, err
		}

		return nil, errors.Wrap(err, "failed to get ruleset")
	}

	if ruleset.Type != typ {
		return nil, ErrTypeMismatch
	}

	res, err := ruleset.Eval(params)
	if err != nil {
		if err == rule.ErrNoMatch {
			return nil, err
		}

		return nil, errors.Wrap(err, "failed to evaluate ruleset")
	}

	if res.Type != typ {
		return nil, ErrTypeMismatch
	}

	return res, nil
}

// GetString evaluates the ruleset associated with key and returns the result as a string.
func (e *Engine) GetString(ctx context.Context, key string, params rule.Params) (string, error) {
	res, err := e.get(ctx, "string", key, params)
	if err != nil {
		return "", err
	}

	return res.Data, nil
}

// GetBool evaluates the ruleset associated with key and returns the result as a bool.
func (e *Engine) GetBool(ctx context.Context, key string, params rule.Params) (bool, error) {
	res, err := e.get(ctx, "bool", key, params)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(res.Data)
}

// GetInt64 evaluates the ruleset associated with key and returns the result as an int64.
func (e *Engine) GetInt64(ctx context.Context, key string, params rule.Params) (int64, error) {
	res, err := e.get(ctx, "int64", key, params)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(res.Data, 10, 64)
}

// GetFloat64 evaluates the ruleset associated with key and returns the result as a float64.
func (e *Engine) GetFloat64(ctx context.Context, key string, params rule.Params) (float64, error) {
	res, err := e.get(ctx, "float64", key, params)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(res.Data, 64)
}

// LoadStruct takes a pointer to struct and params and loads rulesets into fields
// tagged with the "ruleset" struct tag.
func (e *Engine) LoadStruct(ctx context.Context, to interface{}, params rule.Params) error {
	b := backend.Func("regula", func(ctx context.Context, key string) ([]byte, error) {
		ruleset, err := e.getter.Get(ctx, key)
		if err != nil {
			if err == ErrRulesetNotFound {
				return nil, backend.ErrNotFound
			}

			return nil, errors.Wrap(err, "failed to get ruleset")
		}

		val, err := ruleset.Eval(params)
		if err != nil {
			return nil, err
		}

		return []byte(val.Data), nil
	})

	l := confita.NewLoader(b)
	l.Tag = "ruleset"

	return l.Load(ctx, to)
}

// A Getter allows a ruleset to be retrieved.
type Getter interface {
	// Get returns the ruleset associated with the given key.
	// If no ruleset is found for a given key, the implementation must return ErrRulesetNotFound.
	Get(ctx context.Context, key string) (*rule.Ruleset, error)
	// GetVersion returns the ruleset associated with the given key and version.
	// If no ruleset is found for a given key, the implementation must return ErrRulesetNotFound.
	GetVersion(ctx context.Context, key string, version string) (*rule.Ruleset, error)
}

// MemoryGetter is an in-memory getter which stores rulesets in a map.
type MemoryGetter struct {
	Rulesets map[MemoryGetterKey]*rule.Ruleset
}

// MemoryGetterKey is used by MemoryGetter to identify a ruleset and its version within a map.
type MemoryGetterKey struct {
	Path, Version string
}

// Get returns the selected ruleset from memory or returns ErrRulesetNotFound.
func (g *MemoryGetter) Get(ctx context.Context, path string) (*rule.Ruleset, error) {
	return g.GetVersion(ctx, path, "latest")
}

// GetVersion returns the selected ruleset from memory using the given path and version or returns ErrRulesetNotFound.
func (g *MemoryGetter) GetVersion(_ context.Context, path, version string) (*rule.Ruleset, error) {
	r, ok := g.Rulesets[MemoryGetterKey{
		Version: version,
		Path:    path,
	}]
	if !ok {
		return nil, ErrRulesetNotFound
	}

	return r, nil
}
