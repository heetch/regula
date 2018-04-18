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
	"github.com/heetch/rules-engine/client"
	"github.com/heetch/rules-engine/rule"
	"github.com/heetch/rules-engine/store"
	"github.com/pkg/errors"
)

const (
	timeout = 5 * time.Second
)

var (
	_ client.Client = new(Client)
)

// Client queries etcd for rulesets, holds them in memory and keeps them synchronized with the remote store.
// This client is temporary and will be replaced by a version using the API server instead of directly talking
// to the database.
type Client struct {
	sync.RWMutex

	logger    *log.Logger
	keyPrefix string
	rulesets  map[string]*rule.Ruleset
	wcancel   func()
	wg        sync.WaitGroup
}

// Options to customize the Client.
type Options struct {
	Prefix string
	Logger *log.Logger
}

// NewClient takes a connected etcd client, fetches all the rulesets under the given prefix and stores them in the returned client.
// It runs a watcher that keeps the in memory rulesets synchronized with changes on etcd.
// Any leading slash found on keyPrefix is removed and a trailing slash is added automatically before usage.
func NewClient(cli *clientv3.Client, opts Options) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := Client{
		keyPrefix: path.Join(strings.TrimLeft(opts.Prefix, "/"), "/"),
		rulesets:  make(map[string]*rule.Ruleset),
		logger:    opts.Logger,
	}

	if client.logger == nil {
		client.logger = log.New(ioutil.Discard, "", 0)
	}

	resp, err := cli.Get(ctx, client.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve keys under the given keyPrefix")
	}

	for _, kv := range resp.Kvs {
		err := client.storeRuleset(string(kv.Key), kv.Value)
		if err != nil {
			return nil, err
		}
	}

	wctx, wcancel := context.WithCancel(context.Background())
	wch := clientv3.NewWatcher(cli).Watch(wctx, client.keyPrefix, clientv3.WithPrefix())

	client.wcancel = wcancel

	client.wg.Add(1)
	go func() {
		defer client.wg.Done()

		client.watchRulesets(wch)
	}()

	return &client, nil
}

// Get returns a memory stored ruleset based on a given key.
// No network round trip is performed during this call.
func (c *Client) Get(ctx context.Context, key string) (*rule.Ruleset, error) {
	c.RLock()
	defer c.RUnlock()

	rs, ok := c.rulesets[path.Join("/", key)]
	if !ok {
		return nil, client.ErrRulesetNotFound
	}

	return rs, nil
}

func (c *Client) watchRulesets(ch clientv3.WatchChan) {
	for wresp := range ch {
		c.logger.Printf("Synchronizing %d ruleset events\n", len(wresp.Events))
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				err := c.storeRuleset(string(ev.Kv.Key), ev.Kv.Value)
				if err != nil {
					c.logger.Printf("Can't decode ruleset %s from etcd store\n", ev.Kv.Key)
				} else {
					c.logger.Printf("New ruleset %s synchronized from etcd store\n", ev.Kv.Key)
				}
			case mvccpb.DELETE:
				k := string(ev.Kv.Key)
				c.deleteRuleset(k)
				c.logger.Printf("Ruleset %s deleted\n", k)
			}
		}
	}
}

func (c *Client) storeRuleset(key string, value []byte) error {
	c.Lock()
	defer c.Unlock()

	var re store.RulesetEntry
	if err := json.Unmarshal(value, &re); err != nil {
		return err
	}

	k := strings.TrimLeft(key, c.keyPrefix)
	c.rulesets[path.Join("/", k)] = re.Ruleset
	return nil
}

func (c *Client) deleteRuleset(key string) {
	c.Lock()
	k := strings.TrimLeft(key, c.keyPrefix)
	delete(c.rulesets, path.Join("/", k))
	c.Unlock()
}

// Close the etcd watcher.
func (c *Client) Close() {
	c.wcancel()
	c.wg.Wait()
}
