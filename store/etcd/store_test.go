package etcd_test

import (
	"context"
	"encoding/json"
	ppath "path"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/rules-engine/store"
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
		Client:    cli,
		Namespace: "rules-engine-store-tests",
	}

	return &s, func() {
		cli.Delete(context.Background(), s.Namespace, clientv3.WithPrefix())
		cli.Close()
	}
}

func createEntry(t *testing.T, s *etcd.Store, path string, e *store.RulesetEntry) string {
	v, err := json.Marshal(e)
	require.NoError(t, err)

	_, err = s.Client.KV.Put(context.Background(), ppath.Join(s.Namespace, path), string(v))
	require.NoError(t, err)

	return path
}

func TestList(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	t.Run("Root", func(t *testing.T) {
		createEntry(t, s, "a", &store.RulesetEntry{Path: "a"})
		createEntry(t, s, "a", &store.RulesetEntry{Path: "a"})
		createEntry(t, s, "b", &store.RulesetEntry{Path: "b"})
		createEntry(t, s, "c", &store.RulesetEntry{Path: "c"})

		paths := []string{"a", "b", "c"}

		entries, err := s.List(context.Background(), "")
		require.NoError(t, err)
		require.Len(t, entries, len(paths))
		for _, e := range entries {
			require.Contains(t, paths, e.Path)
		}
	})

	t.Run("Prefix", func(t *testing.T) {
		createEntry(t, s, "x", &store.RulesetEntry{Path: "x"})
		createEntry(t, s, "xx", &store.RulesetEntry{Path: "xx"})
		createEntry(t, s, "x/1", &store.RulesetEntry{Path: "x/1"})
		createEntry(t, s, "x/2", &store.RulesetEntry{Path: "x/2"})

		paths := []string{"x", "xx", "x/1", "x/2"}

		entries, err := s.List(context.Background(), "x")
		require.NoError(t, err)
		require.Len(t, entries, len(paths))
		for _, e := range entries {
			require.Contains(t, paths, e.Path)
		}
	})
}

func TestOne(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	createEntry(t, s, "a", &store.RulesetEntry{Path: "a"})
	createEntry(t, s, "b", &store.RulesetEntry{Path: "b"})
	createEntry(t, s, "c", &store.RulesetEntry{Path: "c"})
	createEntry(t, s, "abc", &store.RulesetEntry{Path: "abc"})
	createEntry(t, s, "abcd", &store.RulesetEntry{Path: "abcd"})

	t.Run("OK", func(t *testing.T) {
		path := "a"

		entry, err := s.One(context.Background(), path)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
	})

	t.Run("NOK - path doesn't exist", func(t *testing.T) {
		path := "aa"

		_, err := s.One(context.Background(), path)
		require.Error(t, err)
		require.EqualError(t, err, store.ErrNotFound.Error())
	})

	t.Run("NOK - path exists but it's not a leaf", func(t *testing.T) {
		path := "ab"

		_, err := s.One(context.Background(), path)
		require.Error(t, err)
		require.EqualError(t, err, store.ErrNotFound.Error())
	})
}
