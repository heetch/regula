package etcd_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"testing"
	"time"

	"github.com/heetch/rules-engine/store"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/rules-engine/store/etcd"
	"github.com/stretchr/testify/require"
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"localhost:2379", "etcd:2379"}
)

func newEtcdStore(t *testing.T) (*etcd.Store, func()) {
	t.Helper()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	s := etcd.Store{
		Logger:    log.New(ioutil.Discard, "", 0),
		Client:    cli,
		Namespace: "rules-engine-store-tests",
	}

	return &s, func() {
		cli.Delete(context.Background(), s.Namespace, clientv3.WithPrefix())
		cli.Close()
	}
}

func createEntry(t *testing.T, s *etcd.Store, key string, e *store.RulesetEntry) string {
	v, err := json.Marshal(e)
	require.NoError(t, err)

	_, err = s.Client.KV.Put(context.Background(), path.Join(s.Namespace, key), string(v))
	require.NoError(t, err)

	return key
}

func TestAll(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	createEntry(t, s, "a", &store.RulesetEntry{Name: "a"})
	createEntry(t, s, "b", &store.RulesetEntry{Name: "b"})
	createEntry(t, s, "c", &store.RulesetEntry{Name: "c"})

	keys := []string{"a", "b", "c"}

	entries, err := s.All(context.Background())
	require.NoError(t, err)
	require.Len(t, entries, len(keys))
	for _, e := range entries {
		require.Contains(t, keys, e.Name)
	}
}
