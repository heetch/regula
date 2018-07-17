package etcd_test

import (
	"context"
	ppath "path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula/rule"
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

func createRuleset(t *testing.T, s *etcd.Store, path string, r *rule.Ruleset) {
	_, err := s.Put(context.Background(), path, r)
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	t.Run("Root", func(t *testing.T) {
		createRuleset(t, s, "a", nil)
		createRuleset(t, s, "a", nil)
		createRuleset(t, s, "b", nil)
		createRuleset(t, s, "c", nil)

		paths := []string{"a", "a", "b", "c"}

		entries, err := s.List(context.Background(), "")
		require.NoError(t, err)
		require.Len(t, entries, len(paths))
		for _, e := range entries {
			require.Contains(t, paths, e.Path)
		}
	})

	t.Run("Prefix", func(t *testing.T) {
		createRuleset(t, s, "x", nil)
		createRuleset(t, s, "xx", nil)
		createRuleset(t, s, "x/1", nil)
		createRuleset(t, s, "x/2", nil)

		paths := []string{"x", "xx", "x/1", "x/2"}

		entries, err := s.List(context.Background(), "x")
		require.NoError(t, err)
		require.Len(t, entries, len(paths))
		for _, e := range entries {
			require.Contains(t, paths, e.Path)
		}
	})
}

func TestLatest(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	oldRse, _ := rule.NewBoolRuleset(
		rule.New(
			rule.True(),
			rule.ReturnsBool(true),
		),
	)

	newRse, _ := rule.NewStringRuleset(
		rule.New(
			rule.True(),
			rule.ReturnsString("success"),
		),
	)

	createRuleset(t, s, "a", oldRse)
	// sleep 1 second because bson.ObjectID is based on unix timestamp (second).
	time.Sleep(time.Second)
	createRuleset(t, s, "a", newRse)
	createRuleset(t, s, "b", nil)
	createRuleset(t, s, "c", nil)
	createRuleset(t, s, "abc", nil)
	createRuleset(t, s, "abcd", nil)

	t.Run("OK - several versions of a ruleset", func(t *testing.T) {
		path := "a"

		entry, err := s.Latest(context.Background(), path)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.Equal(t, newRse, entry.Ruleset)
	})

	t.Run("OK - only one version of a ruleset", func(t *testing.T) {
		var exp store.RulesetEntry
		path := "b"

		entry, err := s.Latest(context.Background(), path)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.Equal(t, exp.Ruleset, entry.Ruleset)
	})

	t.Run("NOK - path doesn't exist", func(t *testing.T) {
		path := "aa"

		_, err := s.Latest(context.Background(), path)
		require.Error(t, err)
		require.EqualError(t, err, store.ErrNotFound.Error())
	})

	t.Run("NOK - path exists but it's not a ruleset", func(t *testing.T) {
		path := "ab"

		_, err := s.Latest(context.Background(), path)
		require.Error(t, err)
		require.EqualError(t, err, store.ErrNotFound.Error())
	})
}

func TestOneByVersion(t *testing.T) {
	s, cleanup := newEtcdStore(t)
	defer cleanup()

	oldRse, _ := rule.NewBoolRuleset(
		rule.New(
			rule.True(),
			rule.ReturnsBool(true),
		),
	)

	newRse, _ := rule.NewStringRuleset(
		rule.New(
			rule.True(),
			rule.ReturnsString("success"),
		),
	)

	createRuleset(t, s, "a", oldRse)
	entry, err := s.Latest(context.Background(), "a")
	require.NoError(t, err)
	version := entry.Version

	// sleep 1 second because bson.ObjectID is based on unix timestamp (second).
	time.Sleep(time.Second)
	createRuleset(t, s, "a", newRse)
	createRuleset(t, s, "abc", nil)

	t.Run("OK", func(t *testing.T) {
		path := "a"

		entry, err := s.OneByVersion(context.Background(), path, version)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.Equal(t, version, entry.Version)
		require.Equal(t, oldRse, entry.Ruleset)
	})

	t.Run("NOK - path doesn't exist", func(t *testing.T) {
		path := "a"

		_, err := s.OneByVersion(context.Background(), path, "123version")
		require.Error(t, err)
		require.EqualError(t, err, store.ErrNotFound.Error())
	})

	t.Run("NOK - path exists but it's not a ruleset", func(t *testing.T) {
		path := "ab"

		_, err := s.OneByVersion(context.Background(), path, "123version")
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
		createRuleset(t, s, "aa", nil)
		wg.Wait()
	})
}
