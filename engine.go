package regula

import (
	"context"
	"strconv"
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

	var (
		result *EvalResult
		err    error
	)

	if cfg.Version != "" {
		result, err = e.evaluator.EvalVersion(ctx, path, cfg.Version, params)
	} else {
		result, err = e.evaluator.Eval(ctx, path, params)
	}
	if err != nil {
		if err == rerrors.ErrRulesetNotFound || err == rerrors.ErrNoMatch {
			return nil, err
		}
		return nil, errors.Wrap(err, "failed to evaluate ruleset")
	}

	if result.Value.Type != typ {
		return nil, rerrors.ErrTypeMismatch
	}

	return result, nil
}

// GetString evaluates a ruleset and returns the result as a string.
func (e *Engine) GetString(ctx context.Context, path string, params rule.Params, opts ...Option) (string, error) {
	res, err := e.get(ctx, "string", path, params, opts...)
	if err != nil {
		return "", err
	}

	return res.ToString()
}

// GetBool evaluates a ruleset and returns the result as a bool.
func (e *Engine) GetBool(ctx context.Context, path string, params rule.Params, opts ...Option) (bool, error) {
	res, err := e.get(ctx, "bool", path, params, opts...)
	if err != nil {
		return false, err
	}

	return res.ToBool()
}

// GetInt64 evaluates a ruleset and returns the result as an int64.
func (e *Engine) GetInt64(ctx context.Context, path string, params rule.Params, opts ...Option) (int64, error) {
	res, err := e.get(ctx, "int64", path, params, opts...)
	if err != nil {
		return 0, err
	}

	return res.ToInt64()
}

// GetFloat64 evaluates a ruleset and returns the result as a float64.
func (e *Engine) GetFloat64(ctx context.Context, path string, params rule.Params, opts ...Option) (float64, error) {
	res, err := e.get(ctx, "float64", path, params, opts...)
	if err != nil {
		return 0, err
	}

	return res.ToFloat64()
}

// LoadStruct takes a pointer to struct and params and loads rulesets into fields
// tagged with the "ruleset" struct tag.
func (e *Engine) LoadStruct(ctx context.Context, to interface{}, params rule.Params) error {
	b := backend.Func("regula", func(ctx context.Context, path string) ([]byte, error) {
		res, err := e.evaluator.Eval(ctx, path, params)
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

// An Evaluator provides methods to evaluate rulesets from any location.
// Long running implementations must listen to the given context for timeout and cancelation.
type Evaluator interface {
	// Eval evaluates a ruleset using the given params.
	// If no ruleset is found for a given path, the implementation must return rerrors.ErrRulesetNotFound.
	Eval(ctx context.Context, path string, params rule.Params) (*EvalResult, error)
	// EvalVersion evaluates a specific version of a ruleset using the given params.
	// If no ruleset is found for a given path, the implementation must return rerrors.ErrRulesetNotFound.
	EvalVersion(ctx context.Context, path string, version string, params rule.Params) (*EvalResult, error)
}

// EvalResult is the product of an evaluation. It contains the value generated as long as some metadata.
type EvalResult struct {
	// Result of the evaluation
	Value *rule.Value
	// Version of the ruleset that generated this value
	Version string
}

func (e *EvalResult) ToString() (string, error) {
	return e.Value.Data, nil
}

func (e *EvalResult) ToInt64() (int64, error) {
	return strconv.ParseInt(e.Value.Data, 10, 64)
}

func (e *EvalResult) ToFloat64() (float64, error) {
	return strconv.ParseFloat(e.Value.Data, 64)
}

func (e *EvalResult) ToBool() (bool, error) {
	return strconv.ParseBool(e.Value.Data)
}

// RulesetBuffer can hold a group of rulesets in memory and can be used as an evaluator.
// It is safe for concurrent use.
type RulesetBuffer struct {
	rw       sync.RWMutex
	rulesets map[string][]*rulesetInfo
}

// NewRulesetBuffer creates a ready to use RulesetBuffer.
func NewRulesetBuffer() *RulesetBuffer {
	return &RulesetBuffer{
		rulesets: make(map[string][]*rulesetInfo),
	}
}

type rulesetInfo struct {
	path, version string
	r             *Ruleset
}

// Add adds the given ruleset version to a list for a specific path.
// The last added ruleset is treated as the latest version.
func (b *RulesetBuffer) Add(path, version string, r *Ruleset) {
	b.rw.Lock()
	b.rulesets[path] = append(b.rulesets[path], &rulesetInfo{path, version, r})
	b.rw.Unlock()
}

// Latest returns the latest version of a ruleset.
func (b *RulesetBuffer) Latest(path string) (*Ruleset, string, error) {
	b.rw.RLock()
	defer b.rw.RUnlock()

	l, ok := b.rulesets[path]
	if !ok || len(l) == 0 {
		return nil, "", rerrors.ErrRulesetNotFound
	}

	return l[len(l)-1].r, l[len(l)-1].version, nil
}

// GetVersion returns a ruleset associated with the given path and version.
func (b *RulesetBuffer) GetVersion(path, version string) (*Ruleset, error) {
	b.rw.RLock()
	defer b.rw.RUnlock()

	ri, err := b.getVersion(path, version)
	if err != nil {
		return nil, err
	}

	return ri.r, nil
}

// Eval evaluates the latest added ruleset or returns rerrors.ErrRulesetNotFound if not found.
func (b *RulesetBuffer) Eval(ctx context.Context, path string, params rule.Params) (*EvalResult, error) {
	b.rw.RLock()
	defer b.rw.RUnlock()

	l, ok := b.rulesets[path]
	if !ok || len(l) == 0 {
		return nil, rerrors.ErrRulesetNotFound
	}

	ri := l[len(l)-1]
	v, err := ri.r.Eval(params)
	if err != nil {
		return nil, err
	}

	return &EvalResult{
		Value:   v,
		Version: ri.version,
	}, nil
}

func (b *RulesetBuffer) getVersion(path, version string) (*rulesetInfo, error) {
	l, ok := b.rulesets[path]
	if !ok || len(l) == 0 {
		return nil, rerrors.ErrRulesetNotFound
	}

	for _, ri := range l {
		if ri.version == version {
			return ri, nil
		}
	}

	return nil, rerrors.ErrRulesetNotFound
}

// EvalVersion evaluates the selected ruleset version or returns rerrors.ErrRulesetNotFound if not found.
func (b *RulesetBuffer) EvalVersion(ctx context.Context, path, version string, params rule.Params) (*EvalResult, error) {
	b.rw.RLock()
	defer b.rw.RUnlock()

	ri, err := b.getVersion(path, version)
	if err != nil {
		return nil, err
	}

	v, err := ri.r.Eval(params)
	if err != nil {
		return nil, err
	}

	return &EvalResult{
		Value:   v,
		Version: ri.version,
	}, nil
}
