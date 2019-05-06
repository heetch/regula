package mock

import (
	"context"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
)

// Ensure RulesetService implements store.RulesetService.
var _ store.RulesetService = new(RulesetService)

// RulesetService mocks the store.RulesetService interface.
type RulesetService struct {
	GetCount         int
	GetFn            func(ctx context.Context, path, version string) (*store.RulesetEntry, error)
	ListCount        int
	ListFn           func(context.Context, string, *store.ListOptions) (*store.RulesetEntries, error)
	WatchCount       int
	WatchFn          func(context.Context, string, string) (*store.RulesetEvents, error)
	PutCount         int
	PutFn            func(context.Context, string) (*store.RulesetEntry, error)
	EvalCount        int
	EvalFn           func(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error)
	EvalVersionCount int
	EvalVersionFn    func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
	PatchCount       int
	PatchFn          func(context.Context, string) (*store.RulesetEntry, error)
}

// Get runs GetFn if provided and increments GetCount when invoked.
func (s *RulesetService) Get(ctx context.Context, path, version string) (*store.RulesetEntry, error) {
	s.GetCount++

	if s.GetFn != nil {
		return s.GetFn(ctx, path, version)
	}

	return nil, nil
}

// List runs ListFn if provided and increments ListCount when invoked.
func (s *RulesetService) List(ctx context.Context, prefix string, opt *store.ListOptions) (*store.RulesetEntries, error) {
	s.ListCount++

	if s.ListFn != nil {
		return s.ListFn(ctx, prefix, opt)
	}

	return nil, nil
}

// Watch runs WatchFn if provided and increments WatchCount when invoked.
func (s *RulesetService) Watch(ctx context.Context, prefix, revision string) (*store.RulesetEvents, error) {
	s.WatchCount++

	if s.WatchFn != nil {
		return s.WatchFn(ctx, prefix, revision)
	}

	return nil, nil
}

// Put runs PutFn if provided and increments PutCount when invoked.
func (s *RulesetService) Put(ctx context.Context, path string, ruleset *regula.Ruleset) (*store.RulesetEntry, error) {
	s.PutCount++

	if s.PutFn != nil {
		return s.PutFn(ctx, path)
	}
	return nil, nil
}

// Eval runs EvalFn if provided and increments EvalCount when invoked.
func (s *RulesetService) Eval(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error) {
	s.EvalCount++

	if s.EvalFn != nil {
		return s.EvalFn(ctx, path, params)
	}
	return nil, nil
}

// EvalVersion runs EvalVersionFn if provided and increments EvalVersionCount when invoked.
func (s *RulesetService) EvalVersion(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
	s.EvalVersionCount++

	if s.EvalVersionFn != nil {
		return s.EvalVersionFn(ctx, path, version, params)
	}
	return nil, nil
}

// Patch runs PatchFn if provided and increments PatchCount when invoked.
//
func (s *RulesetService) Patch(ctx context.Context, path string, ruleset *regula.Ruleset) (*store.RulesetEntry, error) {
	s.PatchCount++

	if s.PatchFn != nil {
		return s.PatchFn(ctx, path)
	}
	return nil, nil
}
