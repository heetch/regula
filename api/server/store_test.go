package server

import (
	"context"

	"github.com/heetch/rules-engine/store"
)

var _ store.Store = new(mockStore)

type mockStore struct {
	AllCount int
	AllFn    func(context.Context) ([]store.RulesetEntry, error)
}

func (s *mockStore) All(ctx context.Context) ([]store.RulesetEntry, error) {
	s.AllCount++

	if s.AllFn != nil {
		return s.AllFn(ctx)
	}

	return nil, nil
}
