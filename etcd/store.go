package etcd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
	"github.com/pkg/errors"
)

const (
	timeout = 5 * time.Second
)

// Store is an etcd store that holds rulesets in memory and keeps them synchronized with the remote store.
type Store struct {
	sync.RWMutex

	logger    *log.Logger
	keyPrefix string
	rulesets  map[string]*rule.Ruleset
	wcancel   func()
	wg        sync.WaitGroup
}

// Options to customize the Store
type Options struct {
	Prefix string
	Logger *log.Logger
}

// NewStore takes a connected etcd client, fetches all the rulesets under the given prefix and stores them in the returned store.
// It runs a watcher that keeps the in memory rulesets synchronized with changes on etcd.
// Any leading slash found on keyPrefix is removed and a trailing slash is added automatically before usage.
func NewStore(client *clientv3.Client, opts Options) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	store := Store{
		keyPrefix: path.Join(strings.TrimLeft(opts.Prefix, "/"), "/"),
		rulesets:  make(map[string]*rule.Ruleset),
		logger:    opts.Logger,
	}

	if store.logger == nil {
		store.logger = log.New(ioutil.Discard, "", 0)
	}

	resp, err := client.Get(ctx, store.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve keys under the given keyPrefix")
	}

	for _, kv := range resp.Kvs {
		err := store.storeRuleset(string(kv.Key), kv.Value)
		if err != nil {
			return nil, err
		}
	}

	wctx, wcancel := context.WithCancel(context.Background())
	wch := clientv3.NewWatcher(client).Watch(wctx, store.keyPrefix, clientv3.WithPrefix())

	store.wcancel = wcancel

	store.wg.Add(1)
	go func() {
		defer store.wg.Done()

		store.watchRulesets(wch)
	}()

	return &store, nil
}

// Get returns a memory stored ruleset based on a given key.
// No network round trip is performed during this call.
func (s *Store) Get(key string) (*rule.Ruleset, error) {
	s.RLock()
	defer s.RUnlock()

	rs, ok := s.rulesets[path.Join("/", key)]
	if !ok {
		return nil, rules.ErrRulesetNotFound
	}

	return rs, nil
}

func (s *Store) watchRulesets(c clientv3.WatchChan) {
	for wresp := range c {
		s.logger.Printf("Synchronizing %d ruleset events\n", len(wresp.Events))
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				err := s.storeRuleset(string(ev.Kv.Key), ev.Kv.Value)
				if err != nil {
					s.logger.Printf("Can't decode ruleset %s from etcd store\n", ev.Kv.Key)
				} else {
					s.logger.Printf("New ruleset %s synchronized from etcd store\n", ev.Kv.Key)
				}
			case mvccpb.DELETE:
				k := string(ev.Kv.Key)
				s.deleteRuleset(k)
				s.logger.Printf("Ruleset %s deleted\n", k)
			}
		}
	}
}

func (s *Store) storeRuleset(key string, value []byte) error {
	s.Lock()
	defer s.Unlock()

	var rs rule.Ruleset
	if err := json.Unmarshal(value, &rs); err != nil {
		return err
	}

	k := strings.TrimLeft(key, s.keyPrefix)
	s.rulesets[path.Join("/", k)] = &rs
	return nil
}

func (s *Store) deleteRuleset(key string) {
	s.Lock()
	k := strings.TrimLeft(key, s.keyPrefix)
	delete(s.rulesets, path.Join("/", k))
	s.Unlock()
}

// Close the etcd watcher
func (s *Store) Close() {
	s.wcancel()
	s.wg.Wait()
}
