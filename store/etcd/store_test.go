package etcd_test

import (
	"context"
	"fmt"
	"math/rand"
	ppath "path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula"
	"github.com/heetch/regula/store"
	"github.com/heetch/regula/store/etcd"
	"github.com/stretchr/testify/require"
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"localhost:2379", "etcd:2379"}
)

func Init() {
	rand.Seed(time.Now().Unix())
}

func newEtcdStore(t *testing.T) (*etcd.Store, func()) {
	t.Helper()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	s := etcd.Store{
		Client:    cli,
		Namespace: fmt.Sprintf("regula-store-tests-%d", rand.Int()),
	}

	return &s, func() {
		cli.Delete(context.Background(), s.Namespace, clientv3.WithPrefix())
		cli.Close()
	}
}

func createRuleset(t *testing.T, s *etcd.Store, path string, r *regula.Ruleset) {
	_, err := s.Put(context.Background(), path, r)
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdStore(t)
	defer cleanup()

	t.Run("Root", func(t *testing.T) {
		createRuleset(t, s, "c", nil)
		createRuleset(t, s, "a", nil)
		createRuleset(t, s, "a/1", nil)
		createRuleset(t, s, "b", nil)
		createRuleset(t, s, "a", nil)

		paths := []string{"a/1", "a", "a", "b", "c"}

		entries, err := s.List(context.Background(), "")
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})

	t.Run("Prefix", func(t *testing.T) {
		createRuleset(t, s, "x", nil)
		createRuleset(t, s, "xx", nil)
		createRuleset(t, s, "x/1", nil)
		createRuleset(t, s, "x/2", nil)

		paths := []string{"x/1", "x", "x/2", "xx"}

		entries, err := s.List(context.Background(), "x")
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})
}

func TestLatest(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdStore(t)
	defer cleanup()

	oldRse, _ := regula.NewBoolRuleset(
		regula.NewRule(
			regula.True(),
			regula.BoolValue(true),
		),
	)

	newRse, _ := regula.NewStringRuleset(
		regula.NewRule(
			regula.True(),
			regula.StringValue("success"),
		),
	)

	createRuleset(t, s, "a", oldRse)
	// sleep 1 second because ksuid doesn't guarantee the order within the same second since it's based on a 32 bits timestamp (second).
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
		require.Equal(t, err, store.ErrNotFound)
	})

	t.Run("NOK - path exists but it's not a ruleset", func(t *testing.T) {
		path := "ab"

		_, err := s.Latest(context.Background(), path)
		require.Error(t, err)
		require.Equal(t, err, store.ErrNotFound)
	})
}

func TestOneByVersion(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdStore(t)
	defer cleanup()

	oldRse, _ := regula.NewBoolRuleset(
		regula.NewRule(
			regula.True(),
			regula.BoolValue(true),
		),
	)

	newRse, _ := regula.NewStringRuleset(
		regula.NewRule(
			regula.True(),
			regula.StringValue("success"),
		),
	)

	createRuleset(t, s, "a", oldRse)
	entry, err := s.Latest(context.Background(), "a")
	require.NoError(t, err)
	version := entry.Version

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
		require.Equal(t, err, store.ErrNotFound)
	})

	t.Run("NOK - path exists but it's not a ruleset", func(t *testing.T) {
		path := "ab"

		_, err := s.OneByVersion(context.Background(), path, "123version")
		require.Error(t, err)
		require.Equal(t, err, store.ErrNotFound)
	})
}

func TestPut(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdStore(t)
	defer cleanup()

	t.Run("OK", func(t *testing.T) {
		path := "a"
		rs, _ := regula.NewBoolRuleset(
			regula.NewRule(
				regula.True(),
				regula.BoolValue(true),
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
		require.Equal(t, entry.Version, strings.TrimPrefix(string(resp.Kvs[0].Key), ppath.Join(s.Namespace, "a")+"/"))
	})
}

func TestWatch(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdStore(t)
	defer cleanup()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(time.Second)
		createRuleset(t, s, "aa", nil)
		createRuleset(t, s, "ab", nil)
		createRuleset(t, s, "a/1", nil)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := s.Watch(ctx, "a", "")
	require.NoError(t, err)
	require.Len(t, events.Events, 1)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, "aa", events.Events[0].Path)
	require.Equal(t, store.PutEvent, events.Events[0].Type)

	wg.Wait()

	events, err = s.Watch(ctx, "a", events.Revision)
	require.NoError(t, err)
	require.Len(t, events.Events, 2)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, store.PutEvent, events.Events[0].Type)
	require.Equal(t, "ab", events.Events[0].Path)
	require.Equal(t, store.PutEvent, events.Events[1].Type)
	require.Equal(t, "a/1", events.Events[1].Path)
}
