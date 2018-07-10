package server

import (
	"context"

	"github.com/heetch/rules-engine/store"
)

var _ store.Store = new(mockStore)

type mockStore struct {
	ListCount int
	ListFn    func(context.Context, string) ([]store.RulesetEntry, error)
}

func (s *mockStore) List(ctx context.Context, prefix string) ([]store.RulesetEntry, error) {
	s.ListCount++

	if s.ListFn != nil {
		return s.ListFn(ctx, prefix)
	}

	return nil, nil
}
