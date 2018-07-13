package etcd_test

import (
	"context"
	"encoding/json"
	ppath "path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/heetch/rules-engine/rule"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula/store"
	"github.com/heetch/regula/store/etcd"
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
		Namespace: "regula-store-tests",
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

	t.Run("NOK - path exists but it's not a ruleset", func(t *testing.T) {
		path := "ab"

		_, err := s.One(context.Background(), path)
		require.Error(t, err)
		require.EqualError(t, err, store.ErrNotFound.Error())
	})
}

func TestPut(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	t.Run("OK", func(t *testing.T) {
		path := "a"
		rs, _ := rule.NewBoolRuleset(
			rule.New(
				rule.True(),
				rule.ReturnsBool(true),
			),
		)

		entry, err := s.Put(context.Background(), path, rs)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.NotEmpty(t, entry.Version)
		require.Equal(t, rs, entry.Ruleset)

		resp, err := s.Client.Get(context.Background(), ppath.Join(s.Namespace, path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, resp.Count, 1)
		// verify if the path contains the right ruleset version
		require.Equal(t, entry.Version, strings.TrimLeft(string(resp.Kvs[0].Key), ppath.Join(s.Namespace, "a")+"/"))
	})
}

func TestWatch(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	t.Run("Prefix", func(t *testing.T) {
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			events, err := s.Watch(ctx, "a")
			require.NoError(t, err)
			require.Len(t, events, 1)
			require.Equal(t, store.PutEvent, events[0].Type)
		}()

		time.Sleep(1 * time.Second)
		createEntry(t, s, "aa", &store.RulesetEntry{Path: "aa"})
		wg.Wait()
	})
}
