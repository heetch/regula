package etcd

import (
	"context"
	"encoding/base64"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula/api"
	"github.com/pkg/errors"
)

// List returns a list of ruleset paths.
// The listing is paginated and can be customised using the ListOptions type.
func (s *RulesetService) List(ctx context.Context, opt api.ListOptions) (*api.Rulesets, error) {
	var opts []clientv3.OpOption

	var key string

	// only fetch keys
	opts = append(opts, clientv3.WithKeysOnly())

	// if a cursor is specified, decode the key from it and start the request from that key
	if opt.Cursor != "" {
		lastPath, err := base64.URLEncoding.DecodeString(opt.Cursor)
		if err != nil {
			return nil, api.ErrInvalidCursor
		}

		key = s.signaturesPath(string(lastPath))

		opts = append(opts, clientv3.WithRange(clientv3.GetPrefixRangeEnd(s.signaturesPath(""))))
	} else {
		key = s.signaturesPath("")
		opts = append(opts, clientv3.WithPrefix())
	}

	// limit the number of results
	opts = append(opts, clientv3.WithLimit(int64(opt.GetLimit())))

	// fetch signatures, the paths will be computed from signature keys
	resp, err := s.Client.KV.Get(ctx, key, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch signatures")
	}

	rulesets := api.Rulesets{
		Revision: resp.Header.Revision,
	}

	rulesets.Paths = make([]string, 0, len(resp.Kvs))
	for _, pair := range resp.Kvs {
		rulesets.Paths = append(rulesets.Paths, s.pathFromKey("signatures", string(pair.Key)))
	}

	// if there are still paths left, generate a new cursor
	if len(rulesets.Paths) == opt.GetLimit() && resp.More {
		lastPath := rulesets.Paths[len(rulesets.Paths)-1]
		// we want to start immediately after the last key
		rulesets.Cursor = base64.URLEncoding.EncodeToString([]byte(lastPath + "\x00"))
	}

	return &rulesets, nil
}
