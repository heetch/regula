package client

import (
	"context"
	"sync"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
)

// Evaluator can cache rulesets in memory and can be passed to a regula.Engine to evaluate rulesets without
// network round trips. If required, it can watch the server for changes and update its local cache.
type Evaluator struct {
	*regula.RulesetBuffer
	cancel func()
	wg     sync.WaitGroup
}

// NewEvaluator uses the given client to fetch a list of rulesets starting with the given prefix
// and returns an evaluator that holds the results in memory.
// If watch is true, the evaluator will watch for changes on the server and automatically update
// the underlying RulesetBuffer.
// If watch is set to true, the Close method must always be called to gracefully close the watcher.
func NewEvaluator(ctx context.Context, client *Client, prefix string, watch bool) (*Evaluator, error) {
	ls, err := client.Rulesets.List(ctx, prefix, &ListOptions{
		Limit: 100, // TODO(asdine): make it configurable in future releases
	})
	if err != nil {
		return nil, err
	}

	buf := regula.NewRulesetBuffer()

	for _, rs := range ls.Rulesets {
		buf.Add(rs.Path, rs.Version, &rs)
	}

	for ls.Continue != "" {
		ls, err = client.Rulesets.List(ctx, prefix, &ListOptions{
			Limit:    100, // TODO(asdine): make it configurable in future releases
			Continue: ls.Continue,
		})
		if err != nil {
			return nil, err
		}

		for _, rs := range ls.Rulesets {
			buf.Add(rs.Path, rs.Version, &rs)
		}
	}

	ev := Evaluator{
		RulesetBuffer: buf,
	}

	if watch {
		ctx, cancel := context.WithCancel(context.Background())
		ev.cancel = cancel
		ev.wg.Add(1)
		go func() {
			defer ev.wg.Done()

			ch := client.Rulesets.Watch(ctx, prefix, ls.Revision)

			for wr := range ch {
				if wr.Err != nil {
					client.Logger.Error().Err(err).Msg("Watching failed")
				}

				for _, ev := range wr.Events.Events {
					switch ev.Type {
					case api.PutEvent:
						buf.Add(ev.Path, ev.Version, &regula.Ruleset{
							Path:    ev.Path,
							Version: ev.Version,
							Rules:   ev.Rules,
						})
					}
				}
			}
		}()
	}

	return &ev, nil
}

// Close stops the watcher if running.
func (e *Evaluator) Close() error {
	if e.cancel != nil {
		e.cancel()
		e.wg.Wait()
	}

	return nil
}
