package etcd

import (
	"context"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/pkg/errors"
)

// Watch a list of paths for changes and return a list of events. If paths is empty or nil,
// watch all paths. If the revision is negative, watch from the latest revision.
// This method blocks until there is a change if one of the paths or until the context is canceled.
// The given context can be used to limit the watch period or to cancel any running one.
func (s *RulesetService) Watch(ctx context.Context, opt api.WatchOptions) (*api.RulesetEvents, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	revision := opt.Revision

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	if revision > 0 {
		// watch from the next revision
		opts = append(opts, clientv3.WithRev(revision+1))
	}

	var events api.RulesetEvents

	wc := s.Client.Watch(ctx, s.rulesPath("", ""), opts...)
	for {
		select {
		case wresp := <-wc:
			if err := wresp.Err(); err != nil {
				return nil, errors.Wrapf(err, "failed to watch paths: '%#v'", opt.Paths)
			}

			revision = wresp.Header.Revision

			if len(wresp.Events) == 0 {
				continue
			}

			var list []api.RulesetEvent
			for _, ev := range wresp.Events {
				// filter keys that haven't been selected
				if !s.shouldIncludeEvent(ev, opt.Paths) {
					s.Logger.Debug().Str("type", string(ev.Type)).Str("key", string(ev.Kv.Key)).Msg("watch: ignoring event key")
					continue
				}

				// filter event types, keep only PUT events
				if ev.Type != mvccpb.PUT {
					s.Logger.Debug().Str("type", string(ev.Type)).Msg("watch: ignoring event type")
					continue
				}

				var pbrs pb.Rules
				err := proto.Unmarshal(ev.Kv.Value, &pbrs)
				if err != nil {
					s.Logger.Error().Bytes("entry", ev.Kv.Value).Msg("watch: unmarshalling failed, ignoring the event")
					continue
				}

				path, version := s.pathVersionFromKey(string(ev.Kv.Key))

				list = append(list, api.RulesetEvent{
					Type:    api.RulesetPutEvent,
					Path:    path,
					Rules:   rulesFromProtobuf(&pbrs),
					Version: version,
				})
			}

			// None of the events matched the user selection, so continue
			// waiting for more.
			// we continue watching
			if len(list) == 0 {
				continue
			}

			events.Events = list
			events.Revision = revision
			return &events, nil
		case <-ctx.Done():
			events.Timeout = true
			// if we received events but ignored them
			// this function will go on until the context is canceled.
			// we need to return the latest received revision so the
			// caller can start after the filtered events.
			events.Revision = revision
			return &events, ctx.Err()
		}
	}
}

// shouldIncludeEvent reports whether the given event should be included
// in the Watch data for the given paths.
func (s *RulesetService) shouldIncludeEvent(ev *clientv3.Event, paths []string) bool {
	// detect if the event key is found in the paths list
	// or that the paths list is empty
	key := string(ev.Kv.Key)
	key = key[:strings.Index(key, versionSeparator)]
	ok := len(paths) == 0
	for i := 0; i < len(paths) && !ok; i++ {
		if key == s.rulesPath(paths[i], "") {
			ok = true
		}
	}

	return ok
}
