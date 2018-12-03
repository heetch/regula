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
	ListCount        int
	ListFn           func(context.Context, string, int, string) (*store.RulesetEntries, error)
	WatchCount       int
	WatchFn          func(context.Context, string, string) (*store.RulesetEvents, error)
	PutCount         int
	PutFn            func(context.Context, string) (*store.RulesetEntry, error)
	EvalCount        int
	EvalFn           func(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error)
	EvalVersionCount int
	EvalVersionFn    func(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error)
}

// List runs ListFn if provided and increments ListCount when invoked.
func (s *RulesetService) List(ctx context.Context, prefix string, limit int, token string) (*store.RulesetEntries, error) {
	s.ListCount++

	if s.ListFn != nil {
		return s.ListFn(ctx, prefix, limit, token)
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
