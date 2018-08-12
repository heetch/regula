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
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/heetch/regula/store/etcd"
	"github.com/stretchr/testify/require"
)

var (
	_ store.RulesetService = new(etcd.RulesetService)
	_ regula.Evaluator     = new(etcd.RulesetService)
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"localhost:2379", "etcd:2379"}
)

func Init() {
	rand.Seed(time.Now().Unix())
}

func newEtcdRulesetService(t *testing.T) (*etcd.RulesetService, func()) {
	t.Helper()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	s := etcd.RulesetService{
		Client:    cli,
		Namespace: fmt.Sprintf("regula-store-tests-%d/", rand.Int()),
	}

	return &s, func() {
		cli.Delete(context.Background(), s.Namespace, clientv3.WithPrefix())
		cli.Close()
	}
}

func createRuleset(t *testing.T, s *etcd.RulesetService, path string, r *regula.Ruleset) *store.RulesetEntry {
	e, err := s.Put(context.Background(), path, r)
	if err != nil && err != store.ErrNotModified {
		require.NoError(t, err)
	}
	return e
}

func TestList(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	rs, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))

	t.Run("Root", func(t *testing.T) {

		createRuleset(t, s, "c", rs)
		createRuleset(t, s, "a", rs)
		createRuleset(t, s, "a/1", rs)
		createRuleset(t, s, "b", rs)
		createRuleset(t, s, "a", rs)

		paths := []string{"a/1", "a", "b", "c"}

		entries, err := s.List(context.Background(), "")
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})

	t.Run("Prefix", func(t *testing.T) {
		createRuleset(t, s, "x", rs)
		createRuleset(t, s, "xx", rs)
		createRuleset(t, s, "x/1", rs)
		createRuleset(t, s, "x/2", rs)

		paths := []string{"x/1", "x", "x/2", "xx"}

		entries, err := s.List(context.Background(), "x")
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := s.List(context.Background(), "doesntexist")
		require.Equal(t, err, store.ErrNotFound)
	})
}

func TestLatest(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	oldRse, _ := regula.NewStringRuleset(
		rule.New(
			rule.True(),
			rule.StringValue("a"),
		),
	)

	newRse, _ := regula.NewStringRuleset(
		rule.New(
			rule.True(),
			rule.StringValue("b"),
		),
	)

	createRuleset(t, s, "a", oldRse)
	// sleep 1 second because ksuid doesn't guarantee the order within the same second since it's based on a 32 bits timestamp (second).
	time.Sleep(time.Second)
	createRuleset(t, s, "a", newRse)

	rs, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
	createRuleset(t, s, "b", rs)
	createRuleset(t, s, "c", rs)
	createRuleset(t, s, "abc", rs)
	createRuleset(t, s, "abcd", rs)

	t.Run("OK - several versions of a ruleset", func(t *testing.T) {
		path := "a"

		entry, err := s.Latest(context.Background(), path)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.Equal(t, newRse, entry.Ruleset)
	})

	t.Run("OK - only one version of a ruleset", func(t *testing.T) {
		path := "b"

		entry, err := s.Latest(context.Background(), path)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.Equal(t, rs, entry.Ruleset)
	})

	t.Run("NOK - path doesn't exist", func(t *testing.T) {
		path := "aa"

		_, err := s.Latest(context.Background(), path)
		require.Equal(t, err, store.ErrNotFound)
	})

	t.Run("NOK - empty path", func(t *testing.T) {
		path := ""

		_, err := s.Latest(context.Background(), path)
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

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	oldRse, _ := regula.NewStringRuleset(
		rule.New(
			rule.True(),
			rule.StringValue("a"),
		),
	)

	newRse, _ := regula.NewStringRuleset(
		rule.New(
			rule.True(),
			rule.StringValue("b"),
		),
	)

	createRuleset(t, s, "a", oldRse)
	entry, err := s.Latest(context.Background(), "a")
	require.NoError(t, err)
	version := entry.Version

	createRuleset(t, s, "a", newRse)

	rs, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
	createRuleset(t, s, "abc", rs)

	t.Run("OK", func(t *testing.T) {
		path := "a"

		entry, err := s.OneByVersion(context.Background(), path, version)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.Equal(t, version, entry.Version)
		require.Equal(t, oldRse, entry.Ruleset)
	})

	t.Run("NOK", func(t *testing.T) {
		paths := []string{
			"a",  // doesn't exist
			"ab", // exists but not a ruleset
			"",   // empty path
		}

		for _, path := range paths {
			_, err := s.OneByVersion(context.Background(), path, "123version")
			require.Equal(t, err, store.ErrNotFound)
		}
	})
}

func TestPut(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	t.Run("OK", func(t *testing.T) {
		path := "a"
		rs, _ := regula.NewBoolRuleset(
			rule.New(
				rule.True(),
				rule.BoolValue(true),
			),
		)

		entry, err := s.Put(context.Background(), path, rs)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.NotEmpty(t, entry.Version)
		require.Equal(t, rs, entry.Ruleset)

		// verify ruleset creation
		resp, err := s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "entries", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)
		// verify if the path contains the right ruleset version
		require.Equal(t, entry.Version, strings.TrimPrefix(string(resp.Kvs[0].Key), ppath.Join(s.Namespace, "rulesets", "entries", "a")+"/"))

		// verify checksum creation
		resp, err = s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "checksums", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		// verify latest pointer creation
		resp, err = s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "latest", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		// create new version with same ruleset
		entry2, err := s.Put(context.Background(), path, rs)
		require.Equal(t, err, store.ErrNotModified)
		require.Equal(t, entry, entry2)

		// create new version with different ruleset
		rs, _ = regula.NewBoolRuleset(
			rule.New(
				rule.True(),
				rule.BoolValue(false),
			),
		)
		entry2, err = s.Put(context.Background(), path, rs)
		require.NoError(t, err)
		require.NotEqual(t, entry.Version, entry2.Version)
	})

	t.Run("Signatures", func(t *testing.T) {
		path := "b"
		rs1, err := regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
				),
				rule.BoolValue(true),
			),
		)
		require.NoError(t, err)

		_, err = s.Put(context.Background(), path, rs1)
		require.NoError(t, err)

		// same params, different return type
		rs2, err := regula.NewStringRuleset(
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
				),
				rule.StringValue("true"),
			),
		)
		require.NoError(t, err)

		_, err = s.Put(context.Background(), path, rs2)
		require.True(t, store.IsValidationError(err))

		// adding new params
		rs3, err := regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		)
		require.NoError(t, err)

		_, err = s.Put(context.Background(), path, rs3)
		require.True(t, store.IsValidationError(err))

		// changing param types
		rs4, err := regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.StringParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		)
		require.NoError(t, err)

		_, err = s.Put(context.Background(), path, rs4)
		require.True(t, store.IsValidationError(err))

		// adding new rule with different param types
		rs5, err := regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.StringParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.StringParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		)

		_, err = s.Put(context.Background(), path, rs5)
		require.True(t, store.IsValidationError(err))

		// adding new rule with correct param types but less
		rs6, _ := regula.NewBoolRuleset(
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
				),
				rule.BoolValue(true),
			),
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
				),
				rule.BoolValue(true),
			),
		)

		_, err = s.Put(context.Background(), path, rs6)
		require.NoError(t, err)
	})
}

func TestWatch(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(time.Second)

		rs, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))

		createRuleset(t, s, "aa", rs)
		createRuleset(t, s, "ab", rs)
		createRuleset(t, s, "a/1", rs)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := s.Watch(ctx, "a", "")
	require.NoError(t, err)
	require.Len(t, events.Events, 1)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, "aa", events.Events[0].Path)
	require.Equal(t, store.RulesetPutEvent, events.Events[0].Type)

	wg.Wait()

	events, err = s.Watch(ctx, "a", events.Revision)
	require.NoError(t, err)
	require.Len(t, events.Events, 2)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, store.RulesetPutEvent, events.Events[0].Type)
	require.Equal(t, "ab", events.Events[0].Path)
	require.Equal(t, store.RulesetPutEvent, events.Events[1].Type)
	require.Equal(t, "a/1", events.Events[1].Path)
}

func TestEval(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	rs, _ := regula.NewBoolRuleset(
		rule.New(
			rule.Eq(
				rule.StringParam("id"),
				rule.StringValue("123"),
			),
			rule.BoolValue(true),
		),
	)

	entry := createRuleset(t, s, "a", rs)

	t.Run("OK", func(t *testing.T) {
		res, err := s.Eval(context.Background(), "a", regula.Params{
			"id": "123",
		})
		require.NoError(t, err)
		require.Equal(t, entry.Version, res.Version)
		require.Equal(t, rule.BoolValue(true), res.Value)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := s.Eval(context.Background(), "notexists", regula.Params{
			"id": "123",
		})
		require.Equal(t, regula.ErrRulesetNotFound, err)
	})
}

func TestEvalVersion(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	rs, _ := regula.NewBoolRuleset(
		rule.New(
			rule.Eq(
				rule.StringParam("id"),
				rule.StringValue("123"),
			),
			rule.BoolValue(true),
		),
	)

	entry := createRuleset(t, s, "a", rs)

	t.Run("OK", func(t *testing.T) {
		res, err := s.EvalVersion(context.Background(), "a", entry.Version, regula.Params{
			"id": "123",
		})
		require.NoError(t, err)
		require.Equal(t, entry.Version, res.Version)
		require.Equal(t, rule.BoolValue(true), res.Value)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := s.EvalVersion(context.Background(), "b", entry.Version, regula.Params{
			"id": "123",
		})
		require.Equal(t, regula.ErrRulesetNotFound, err)
	})

	t.Run("BadVersion", func(t *testing.T) {
		_, err := s.EvalVersion(context.Background(), "a", "someversion", regula.Params{
			"id": "123",
		})
		require.Equal(t, regula.ErrRulesetNotFound, err)
	})
}
