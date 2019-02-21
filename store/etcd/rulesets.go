package etcd

import (
	"context"
	"path"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula/store"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// versionSeparator separates the path from the version in the entries path in etcd.
// The purpose is to have the same ordering as the others namespace (latest, versions, ...).
const versionSeparator = "!"

// RulesetService manages the rulesets using etcd.
type RulesetService struct {
	Client    *clientv3.Client
	Logger    zerolog.Logger
	Namespace string
}

// Watch the given prefix for anything new.
func (s *RulesetService) Watch(ctx context.Context, prefix string, revision string) (*store.RulesetEvents, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	if i, _ := strconv.ParseInt(revision, 10, 64); i > 0 {
		// watch from the next revision
		opts = append(opts, clientv3.WithRev(i+1))
	}

	wc := s.Client.Watch(ctx, s.entriesPath(prefix, ""), opts...)
	for {
		select {
		case wresp := <-wc:
			if err := wresp.Err(); err != nil {
				return nil, errors.Wrapf(err, "failed to watch prefix: '%s'", prefix)
			}

			if len(wresp.Events) == 0 {
				continue
			}

			events := make([]store.RulesetEvent, len(wresp.Events))
			for i, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					events[i].Type = store.RulesetPutEvent
				default:
					s.Logger.Debug().Str("type", string(ev.Type)).Msg("watch: ignoring event type")
					continue
				}

				var pbrse pb.RulesetEntry
				err := proto.Unmarshal(ev.Kv.Value, &pbrse)
				if err != nil {
					s.Logger.Debug().Bytes("entry", ev.Kv.Value).Msg("watch: unmarshalling failed")
					return nil, errors.Wrap(err, "failed to unmarshal entry")
				}
				events[i].Path = pbrse.Path
				events[i].Ruleset = rulesetFromProtobuf(pbrse.Ruleset)
				events[i].Version = pbrse.Version
			}

			return &store.RulesetEvents{
				Events:   events,
				Revision: strconv.FormatInt(wresp.Header.Revision, 10),
			}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}

// entriesPath returns the path where the rulesets are stored in etcd.
func (s *RulesetService) entriesPath(p, v string) string {
	// If the version parameter is not empty, we concatenate to the path separated by the versionSeparator value.
	if v != "" {
		p += versionSeparator + v
	}
	return path.Join(s.Namespace, "rulesets", "entries", p)
}

// checksumsPath returns the path where the checksums are stored in etcd.
func (s *RulesetService) checksumsPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "checksums", p)
}

// signaturesPath returns the path where the signatures are stored in etcd.
func (s *RulesetService) signaturesPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "signatures", p)
}

// latestRulesetPath returns the path where the latest version of each ruleset is stored in etcd.
func (s *RulesetService) latestRulesetPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "latest", p)
}

// versionsPath returns the path where the versions of each rulesets are stored in etcd.
func (s *RulesetService) versionsPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "versions", p)
}
