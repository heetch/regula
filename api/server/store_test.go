package server

import (
	"context"

	"github.com/heetch/regula"
	"github.com/heetch/regula/store"
)

var _ store.RulesetService = new(mockRulesetService)

type mockRulesetService struct {
	ListCount         int
	ListFn            func(context.Context, string) (*store.RulesetEntries, error)
	LatestCount       int
	LatestFn          func(context.Context, string) (*store.RulesetEntry, error)
	OneByVersionCount int
	OneByVersionFn    func(context.Context, string, string) (*store.RulesetEntry, error)
	WatchCount        int
	WatchFn           func(context.Context, string, string) (*store.RulesetEvents, error)
	PutCount          int
	PutFn             func(context.Context, string) (*store.RulesetEntry, error)
	EvalCount         int
	EvalFn            func(ctx context.Context, path string, params regula.ParamGetter) (*regula.EvalResult, error)
	EvalVersionCount  int
	EvalVersionFn     func(ctx context.Context, path, version string, params regula.ParamGetter) (*regula.EvalResult, error)
}

func (s *mockRulesetService) List(ctx context.Context, prefix string) (*store.RulesetEntries, error) {
	s.ListCount++

	if s.ListFn != nil {
		return s.ListFn(ctx, prefix)
	}

	return nil, nil
}

func (s *mockRulesetService) Latest(ctx context.Context, path string) (*store.RulesetEntry, error) {
	s.LatestCount++

	if s.LatestFn != nil {
		return s.LatestFn(ctx, path)
	}
	return nil, nil
}

func (s *mockRulesetService) OneByVersion(ctx context.Context, path, version string) (*store.RulesetEntry, error) {
	s.OneByVersionCount++

	if s.OneByVersionFn != nil {
		return s.OneByVersionFn(ctx, path, version)
	}
	return nil, nil
}

func (s *mockRulesetService) Watch(ctx context.Context, prefix, revision string) (*store.RulesetEvents, error) {
	s.WatchCount++

	if s.WatchFn != nil {
		return s.WatchFn(ctx, prefix, revision)
	}

	return nil, nil
}

func (s *mockRulesetService) Put(ctx context.Context, path string, ruleset *regula.Ruleset) (*store.RulesetEntry, error) {
	s.PutCount++

	if s.PutFn != nil {
		return s.PutFn(ctx, path)
	}
	return nil, nil
}

func (s *mockRulesetService) Eval(ctx context.Context, path string, params regula.ParamGetter) (*regula.EvalResult, error) {
	s.EvalCount++

	if s.EvalFn != nil {
		return s.EvalFn(ctx, path, params)
	}
	return nil, nil
}

func (s *mockRulesetService) EvalVersion(ctx context.Context, path, version string, params regula.ParamGetter) (*regula.EvalResult, error) {
	s.EvalVersionCount++

	if s.EvalVersionFn != nil {
		return s.EvalVersionFn(ctx, path, version, params)
	}
	return nil, nil
}
