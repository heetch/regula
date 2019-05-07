package etcd

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(time.Second)

		r := rule.New(rule.True(), rule.BoolValue(true))

		createBoolRuleset(t, s, "aa", r)
		createBoolRuleset(t, s, "ab", r)
		createBoolRuleset(t, s, "a/1", r)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := s.Watch(ctx, "a", "")
	require.NoError(t, err)
	require.Len(t, events.Events, 1)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, "aa", events.Events[0].Path)
	require.Equal(t, api.RulesetPutEvent, events.Events[0].Type)

	wg.Wait()

	events, err = s.Watch(ctx, "a", events.Revision)
	require.NoError(t, err)
	require.Len(t, events.Events, 2)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, api.RulesetPutEvent, events.Events[0].Type)
	require.Equal(t, "ab", events.Events[0].Path)
	require.Equal(t, api.RulesetPutEvent, events.Events[1].Type)
	require.Equal(t, "a/1", events.Events[1].Path)

	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		events, err := s.Watch(ctx, "", "")
		require.Equal(t, context.DeadlineExceeded, err)
		require.True(t, events.Timeout)
	})
}
