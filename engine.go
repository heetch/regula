package regula

import (
	"context"
	"strconv"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/pkg/errors"
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
func (e *Engine) get(ctx context.Context, typ, key string, params Params, opts ...Option) (*Value, error) {
	var cfg engineConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	var (
		ruleset *Ruleset
		err     error
	)

	if cfg.Version != "" {
		ruleset, err = e.getter.GetVersion(ctx, key, cfg.Version)
	} else {
		ruleset, err = e.getter.Get(ctx, key)
	}
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
		if err == ErrNoMatch {
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
func (e *Engine) GetString(ctx context.Context, key string, params Params, opts ...Option) (string, error) {
	res, err := e.get(ctx, "string", key, params, opts...)
	if err != nil {
		return "", err
	}

	return res.Data, nil
}

// GetBool evaluates the ruleset associated with key and returns the result as a bool.
func (e *Engine) GetBool(ctx context.Context, key string, params Params, opts ...Option) (bool, error) {
	res, err := e.get(ctx, "bool", key, params, opts...)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(res.Data)
}

// GetInt64 evaluates the ruleset associated with key and returns the result as an int64.
func (e *Engine) GetInt64(ctx context.Context, key string, params Params, opts ...Option) (int64, error) {
	res, err := e.get(ctx, "int64", key, params, opts...)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(res.Data, 10, 64)
}

// GetFloat64 evaluates the ruleset associated with key and returns the result as a float64.
func (e *Engine) GetFloat64(ctx context.Context, key string, params Params, opts ...Option) (float64, error) {
	res, err := e.get(ctx, "float64", key, params, opts...)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(res.Data, 64)
}

// LoadStruct takes a pointer to struct and params and loads rulesets into fields
// tagged with the "ruleset" struct tag.
func (e *Engine) LoadStruct(ctx context.Context, to interface{}, params Params) error {
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

type engineConfig struct {
	Version string
}

// Option is used to customize the engine behaviour.
type Option func(cfg *engineConfig)

// Version is an option used to describe which ruleset version the engine should return.
func Version(version string) Option {
	return func(cfg *engineConfig) {
		cfg.Version = version
	}
}

// A Getter allows a ruleset to be retrieved.
type Getter interface {
	// Get returns the ruleset associated with the given key.
	// If no ruleset is found for a given key, the implementation must return ErrRulesetNotFound.
	Get(ctx context.Context, key string) (*Ruleset, error)
	// GetVersion returns the ruleset associated with the given key and version.
	// If no ruleset is found for a given key, the implementation must return ErrRulesetNotFound.
	GetVersion(ctx context.Context, key string, version string) (*Ruleset, error)
}

// MemoryGetter is an in-memory getter which stores rulesets in a map.
type MemoryGetter struct {
	rulesets map[string][]*rulesetInfo
}

type rulesetInfo struct {
	path, version string
	r             *Ruleset
}

// AddRuleset adds the given ruleset version to a list for a specific path.
// The last added ruleset is treated as the latest version.
func (m *MemoryGetter) AddRuleset(path, version string, r *Ruleset) {
	if m.rulesets == nil {
		m.rulesets = make(map[string][]*rulesetInfo)
	}

	m.rulesets[path] = append(m.rulesets[path], &rulesetInfo{path, version, r})
}

// Get returns the latest added ruleset from memory or returns ErrRulesetNotFound if not found.
func (m *MemoryGetter) Get(ctx context.Context, path string) (*Ruleset, error) {
	if m.rulesets == nil {
		m.rulesets = make(map[string][]*rulesetInfo)
	}

	l, ok := m.rulesets[path]
	if !ok || len(l) == 0 {
		return nil, ErrRulesetNotFound
	}

	return l[len(l)-1].r, nil
}

// GetVersion returns the selected ruleset from memory using the given path and version or returns ErrRulesetNotFound.
func (m *MemoryGetter) GetVersion(ctx context.Context, path, version string) (*Ruleset, error) {
	if m.rulesets == nil {
		m.rulesets = make(map[string][]*rulesetInfo)
	}

	l, ok := m.rulesets[path]
	if !ok || len(l) == 0 {
		return nil, ErrRulesetNotFound
	}

	for _, r := range l {
		if r.version == version {
			return r.r, nil
		}
	}

	return nil, ErrRulesetNotFound
}
