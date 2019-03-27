package regula

import (
	"context"
	"sync"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	rerrors "github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

// Engine is used to evaluate a ruleset against a group of parameters.
// It provides a list of type safe methods to evaluate a ruleset and always returns the expected type to the caller.
// The engine is stateless and relies on the given evaluator to evaluate a ruleset.
// It is safe for concurrent use.
type Engine struct {
	evaluator Evaluator
}

// NewEngine creates an Engine using the given evaluator.
func NewEngine(evaluator Evaluator) *Engine {
	return &Engine{
		evaluator: evaluator,
	}
}

// get evaluates a ruleset and returns the result.
func (e *Engine) get(ctx context.Context, typ, path string, params rule.Params, opts ...Option) (*EvalResult, error) {
	var cfg engineConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	result, err := e.evaluator.Eval(ctx, path, cfg.Version, params)
	if err != nil {
		if err == rerrors.ErrRulesetNotFound || err == rerrors.ErrNoMatch {
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to evaluate ruleset")
	}

	if typ != "" && result.Value.Type != typ {
		return nil, errors.New("type returned by rule doesn't match")
	}

	return result, nil
}

// Get evaluates a ruleset and returns the result.
func (e *Engine) Get(ctx context.Context, path string, params rule.Params, opts ...Option) (*EvalResult, error) {
	return e.get(ctx, "", path, params, opts...)
}

// GetString evaluates a ruleset and returns the result as a string.
func (e *Engine) GetString(ctx context.Context, path string, params rule.Params, opts ...Option) (string, error) {
	res, err := e.get(ctx, "string", path, params, opts...)
	if err != nil {
		return "", err
	}

	return res.Value.ToString()
}

// GetBool evaluates a ruleset and returns the result as a bool.
func (e *Engine) GetBool(ctx context.Context, path string, params rule.Params, opts ...Option) (bool, error) {
	res, err := e.get(ctx, "bool", path, params, opts...)
	if err != nil {
		return false, err
	}

	return res.Value.ToBool()
}

// GetInt64 evaluates a ruleset and returns the result as an int64.
func (e *Engine) GetInt64(ctx context.Context, path string, params rule.Params, opts ...Option) (int64, error) {
	res, err := e.get(ctx, "int64", path, params, opts...)
	if err != nil {
		return 0, err
	}

	return res.Value.ToInt64()
}

// GetFloat64 evaluates a ruleset and returns the result as a float64.
func (e *Engine) GetFloat64(ctx context.Context, path string, params rule.Params, opts ...Option) (float64, error) {
	res, err := e.get(ctx, "float64", path, params, opts...)
	if err != nil {
		return 0, err
	}

	return res.Value.ToFloat64()
}

// LoadStruct takes a pointer to struct and params and loads rulesets into fields
// tagged with the "ruleset" struct tag.
func (e *Engine) LoadStruct(ctx context.Context, to interface{}, params rule.Params) error {
	b := backend.Func("regula", func(ctx context.Context, path string) ([]byte, error) {
		res, err := e.evaluator.Eval(ctx, path, "", params)
		if err != nil {
			if err == rerrors.ErrRulesetNotFound {
				return nil, backend.ErrNotFound
			}

			return nil, err
		}

		return []byte(res.Value.Data), nil
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

// An Evaluator provides a method to evaluate rulesets from any location.
// Long running implementations must listen to the given context for timeout and cancelation.
type Evaluator interface {
	// Eval evaluates a ruleset using the given params. If the version is not empty, the selected version
	// will be used for evaluation. If it's empty, the latest version will be used.
	// If no ruleset is found for a given path, the implementation must return errors.ErrRulesetNotFound.
	Eval(ctx context.Context, path string, version string, params rule.Params) (*EvalResult, error)
}

// EvalResult is the product of an evaluation. It contains the value generated as long as some metadata.
type EvalResult struct {
	// Result of the evaluation
	Value *rule.Value `json:"value"`
	// Version of the ruleset that generated this value
	Version string `json:"version"`
}

// RulesetBuffer can hold a group of rulesets in memory and can be used as an evaluator.
// It is safe for concurrent use.
type RulesetBuffer struct {
	rw       sync.RWMutex
	rulesets map[string]*Ruleset
}

// NewRulesetBuffer creates a ready to use RulesetBuffer.
func NewRulesetBuffer() *RulesetBuffer {
	return &RulesetBuffer{
		rulesets: make(map[string]*Ruleset),
	}
}

// Set the given ruleset to a list for a specific path.
func (b *RulesetBuffer) Set(path string, r *Ruleset) {
	b.rw.Lock()
	b.rulesets[path] = r
	b.rw.Unlock()
}

// Get a ruleset associated with path.
func (b *RulesetBuffer) Get(path string) (*Ruleset, error) {
	b.rw.RLock()
	defer b.rw.RUnlock()

	r, ok := b.rulesets[path]
	if !ok {
		return nil, errors.New("ruleset not found")
	}

	return r, nil
}

// Eval evaluates the selected ruleset version, or the latest one if the version is empty, and returns errors.ErrRulesetNotFound if not found.
func (b *RulesetBuffer) Eval(ctx context.Context, path, version string, params rule.Params) (*EvalResult, error) {
	b.rw.RLock()
	defer b.rw.RUnlock()

	r, ok := b.rulesets[path]
	if !ok || len(r.Versions) == 0 {
		return nil, rerrors.ErrRulesetVersionNotFound
	}

	var (
		v   *rule.Value
		err error
	)

	if version == "" {
		version = r.Versions[len(r.Versions)-1].Version
		v, err = r.Eval(params)
	} else {
		v, err = r.EvalVersion(version, params)
	}

	if err != nil {
		return nil, err
	}

	return &EvalResult{
		Value:   v,
		Version: version,
	}, nil
}
