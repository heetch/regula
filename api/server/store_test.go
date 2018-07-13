package server

import (
	"context"

	"github.com/heetch/regula/store"
)

var _ store.Store = new(mockStore)

type mockStore struct {
	ListCount int
	ListFn    func(context.Context, string) ([]store.RulesetEntry, error)

	OneCount int
	OneFn    func(context.Context, string) (*store.RulesetEntry, error)
}

func (s *mockStore) List(ctx context.Context, prefix string) ([]store.RulesetEntry, error) {
	s.ListCount++

	if s.ListFn != nil {
		return s.ListFn(ctx, prefix)
	}

	return nil, nil
}

func (s *mockStore) One(ctx context.Context, path string) (*store.RulesetEntry, error) {
	s.OneCount++

	if s.OneFn != nil {
		return s.OneFn(ctx, path)
	}

	return nil, nil
}
