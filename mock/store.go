package mock

import (
	"context"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
)

// Ensure RulesetService implements api.RulesetService.
var _ api.RulesetService = new(RulesetService)

// RulesetService mocks the api.RulesetService interface.
type RulesetService struct {
	CreateCount int
	CreateFn    func(ctx context.Context, path string, signature *regula.Signature) error
	GetCount    int
	GetFn       func(ctx context.Context, path, version string) (*regula.Ruleset, error)
	ListCount   int
	ListFn      func(context.Context, api.ListOptions) (*api.Rulesets, error)
	WatchCount  int
	WatchFn     func(context.Context, api.WatchOptions) (*api.RulesetEvents, error)
	PutCount    int
	PutFn       func(context.Context, string, []*rule.Rule) (string, error)
	EvalCount   int
	EvalFn      func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
}

// Create runs CreateFn if provided and increments CreateCount when invoked.
func (s *RulesetService) Create(ctx context.Context, path string, sig *regula.Signature) error {
	s.CreateCount++

	if s.CreateFn != nil {
		return s.CreateFn(ctx, path, sig)
	}

	return nil
}

// Get runs GetFn if provided and increments GetCount when invoked.
func (s *RulesetService) Get(ctx context.Context, path, version string) (*regula.Ruleset, error) {
	s.GetCount++

	if s.GetFn != nil {
		return s.GetFn(ctx, path, version)
	}

	return nil, nil
}

// List runs ListFn if provided and increments ListCount when invoked.
func (s *RulesetService) List(ctx context.Context, opt api.ListOptions) (*api.Rulesets, error) {
	s.ListCount++

	if s.ListFn != nil {
		return s.ListFn(ctx, opt)
	}

	return nil, nil
}

// Watch runs WatchFn if provided and increments WatchCount when invoked.
func (s *RulesetService) Watch(ctx context.Context, opt api.WatchOptions) (*api.RulesetEvents, error) {
	s.WatchCount++

	if s.WatchFn != nil {
		return s.WatchFn(ctx, opt)
	}

	return nil, nil
}

// Put runs PutFn if provided and increments PutCount when invoked.
func (s *RulesetService) Put(ctx context.Context, path string, rules []*rule.Rule) (string, error) {
	s.PutCount++

	if s.PutFn != nil {
		return s.PutFn(ctx, path, rules)
	}
	return "", nil
}

// Eval runs EvalFn if provided and increments EvalCount when invoked.
func (s *RulesetService) Eval(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
	s.EvalCount++

	if s.EvalFn != nil {
		return s.EvalFn(ctx, path, version, params)
	}
	return nil, nil
}
