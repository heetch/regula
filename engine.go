package regula

import (
	"context"
	"strconv"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/pkg/errors"
)

// Engine is used to evaluate a ruleset against a group of parameters.
// It provides a list of type safe methods to evaluate a ruleset and always returns the expected type to the caller.
// The engine is stateless and relies on the given evaluator to evaluate a ruleset.
type Engine struct {
	evaluator Evaluator
}

// NewEngine creates an Engine using the given evaluator.
func NewEngine(evaluator Evaluator) *Engine {
	return &Engine{
		evaluator: evaluator,
	}
}

// Get evaluates a ruleset and returns the result.
func (e *Engine) get(ctx context.Context, typ, path string, params ParamGetter, opts ...Option) (*Value, error) {
	var cfg engineConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	var (
		value *Value
		err   error
	)

	if cfg.Version != "" {
		value, err = e.evaluator.EvalVersion(ctx, path, cfg.Version, params)
	} else {
		value, err = e.evaluator.Eval(ctx, path, params)
	}
	if err != nil {
		if err == ErrRulesetNotFound || err == ErrNoMatch {
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to evaluate ruleset")
	}

	if value.Type != typ {
		return nil, ErrTypeMismatch
	}

	return value, nil
}

// GetString evaluates a ruleset and returns the result as a string.
func (e *Engine) GetString(ctx context.Context, path string, params ParamGetter, opts ...Option) (string, error) {
	res, err := e.get(ctx, "string", path, params, opts...)
	if err != nil {
		return "", err
	}

	return res.Data, nil
}

// GetBool evaluates a ruleset and returns the result as a bool.
func (e *Engine) GetBool(ctx context.Context, path string, params ParamGetter, opts ...Option) (bool, error) {
	res, err := e.get(ctx, "bool", path, params, opts...)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(res.Data)
}

// GetInt64 evaluates a ruleset and returns the result as an int64.
func (e *Engine) GetInt64(ctx context.Context, path string, params ParamGetter, opts ...Option) (int64, error) {
	res, err := e.get(ctx, "int64", path, params, opts...)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(res.Data, 10, 64)
}

// GetFloat64 evaluates a ruleset and returns the result as a float64.
func (e *Engine) GetFloat64(ctx context.Context, path string, params ParamGetter, opts ...Option) (float64, error) {
	res, err := e.get(ctx, "float64", path, params, opts...)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(res.Data, 64)
}

// LoadStruct takes a pointer to struct and params and loads rulesets into fields
// tagged with the "ruleset" struct tag.
func (e *Engine) LoadStruct(ctx context.Context, to interface{}, params ParamGetter) error {
	b := backend.Func("regula", func(ctx context.Context, path string) ([]byte, error) {
		val, err := e.evaluator.Eval(ctx, path, params)
		if err != nil {
			if err == ErrRulesetNotFound {
				return nil, backend.ErrNotFound
			}

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

// An Evaluator provides methods to evaluate rulesets from any location.
// Long running implementations must listen to the given context for timeout and cancelation.
type Evaluator interface {
	// Eval evaluates a ruleset using the given params.
	// If no ruleset is found for a given path, the implementation must return ErrRulesetNotFound.
	Eval(ctx context.Context, path string, params ParamGetter) (*Value, error)
	// EvalVersion evaluates a specific version of a ruleset using the given params.
	// If no ruleset is found for a given path, the implementation must return ErrRulesetNotFound.
	EvalVersion(ctx context.Context, path string, version string, params ParamGetter) (*Value, error)
}

// RulesetBuffer can hold a group of rulesets in memory and can be used as an evaluator.
type RulesetBuffer struct {
	rulesets map[string][]*rulesetInfo
}

type rulesetInfo struct {
	path, version string
	r             *Ruleset
}

// AddRuleset adds the given ruleset version to a list for a specific path.
// The last added ruleset is treated as the latest version.
func (b *RulesetBuffer) AddRuleset(path, version string, r *Ruleset) {
	if b.rulesets == nil {
		b.rulesets = make(map[string][]*rulesetInfo)
	}

	b.rulesets[path] = append(b.rulesets[path], &rulesetInfo{path, version, r})
}

// Eval evaluates the latest added ruleset or returns ErrRulesetNotFound if not found.
func (b *RulesetBuffer) Eval(ctx context.Context, path string, params ParamGetter) (*Value, error) {
	if b.rulesets == nil {
		b.rulesets = make(map[string][]*rulesetInfo)
	}

	l, ok := b.rulesets[path]
	if !ok || len(l) == 0 {
		return nil, ErrRulesetNotFound
	}

	return l[len(l)-1].r.Eval(params)
}

// EvalVersion evaluates the selected ruleset version or returns ErrRulesetNotFound if not found.
func (b *RulesetBuffer) EvalVersion(ctx context.Context, path string, version string, params ParamGetter) (*Value, error) {
	if b.rulesets == nil {
		b.rulesets = make(map[string][]*rulesetInfo)
	}

	l, ok := b.rulesets[path]
	if !ok || len(l) == 0 {
		return nil, ErrRulesetNotFound
	}

	for _, r := range l {
		if r.version == version {
			return r.r.Eval(params)
		}
	}

	return nil, ErrRulesetNotFound
}
