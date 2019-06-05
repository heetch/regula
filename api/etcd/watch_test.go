package etcd

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		paths    []string
		expected []string
	}{
		{"no paths", nil, []string{"a", "b", "c"}},
		{"existing paths", []string{"a", "c"}, []string{"a", "c"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, cleanup := newEtcdRulesetService(t)
			defer cleanup()

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()

				// wait enought time so that the other goroutine had the time to run the watch method
				// before writing data to the database.
				time.Sleep(time.Second)

				r := rule.New(rule.True(), rule.BoolValue(true))

				createBoolRuleset(t, s, "a", r)
				createBoolRuleset(t, s, "b", r)
				createBoolRuleset(t, s, "c", r)
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var events api.RulesetEvents
			var rev int64
			var watchCount int
			for len(events.Events) != len(test.expected) && watchCount < 4 {
				evs, err := s.Watch(ctx, api.WatchOptions{Paths: test.paths, Revision: rev})
				if err != nil {
					if err != nil {
						if err == context.DeadlineExceeded {
							t.Errorf("timed out waiting for expected events")
						} else {
							t.Errorf("unexpected error from watcher: %v", err)
						}
						break
					}
					break
				}
				assert.True(t, len(evs.Events) > 0)
				assert.NotEmpty(t, evs.Revision)
				rev = evs.Revision
				events.Events = append(events.Events, evs.Events...)
				watchCount++
			}

			wg.Wait()

			var foundCount int
			for _, ev := range events.Events {
				for _, p := range test.expected {
					if ev.Path == p {
						foundCount++
						break
					}
				}
			}
			require.Equal(t, len(test.expected), foundCount)
		})

	}

	t.Run("timeout", func(t *testing.T) {
		s, cleanup := newEtcdRulesetService(t)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		events, err := s.Watch(ctx, api.WatchOptions{})
		require.Equal(t, context.DeadlineExceeded, err)
		require.True(t, events.Timeout)
	})
}
