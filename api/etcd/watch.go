package etcd

import (
	"context"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/pkg/errors"
)

// Watch a list of paths for changes and return a list of events. If the list is empty or nil,
// watch all paths. If the revision is empty, watch from the latest revision.
func (s *RulesetService) Watch(ctx context.Context, paths []string, revision string) (*api.RulesetEvents, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	rev, _ := strconv.ParseInt(revision, 10, 64)
	if rev > 0 {
		// watch from the next revision
		opts = append(opts, clientv3.WithRev(rev+1))
	}

	var events api.RulesetEvents

	wc := s.Client.Watch(ctx, s.rulesPath("", ""), opts...)
	for {
		select {
		case wresp := <-wc:
			if err := wresp.Err(); err != nil {
				return nil, errors.Wrapf(err, "failed to watch paths: '%#v'", paths)
			}

			rev = wresp.Header.Revision

			if len(wresp.Events) == 0 {
				continue
			}

			var list []api.RulesetEvent
			for i, ev := range wresp.Events {
				// detect if the event key is found in the paths list
				// or that the paths list is empty
				ok := len(paths) == 0
				for i := 0; i < len(paths) && !ok; i++ {
					if string(ev.Kv.Key) == s.rulesPath(paths[i], "") {
						ok = true
					}
				}

				// filter keys that haven't been selected
				if !ok {
					s.Logger.Debug().Str("type", string(ev.Type)).Str("key", string(ev.Kv.Key)).Msg("watch: ignoring event key")
					continue
				}

				// filter event types, keep only PUT events
				if ev.Type != mvccpb.PUT {
					s.Logger.Debug().Str("type", string(ev.Type)).Msg("watch: ignoring event type")
					continue
				}

				list[i].Type = api.RulesetPutEvent

				var pbrs pb.Rules
				err := proto.Unmarshal(ev.Kv.Value, &pbrs)
				if err != nil {
					s.Logger.Debug().Bytes("entry", ev.Kv.Value).Msg("watch: unmarshalling failed")
					return nil, errors.Wrap(err, "failed to unmarshal entry")
				}
				path, version := s.pathVersionFromKey(string(ev.Kv.Key))
				list[i].Path = path
				list[i].Rules = rulesFromProtobuf(&pbrs)
				list[i].Version = version
			}

			events.Events = list
			events.Revision = strconv.FormatInt(rev, 10)
			return &events, nil
		case <-ctx.Done():
			events.Timeout = true
			// if we received events but ignored them
			// this function will go on until the context is canceled.
			// we need to return the latest received revision so the
			// caller can start after the filtered events.
			events.Revision = strconv.FormatInt(rev, 10)
			return &events, ctx.Err()
		}
	}
}
