package etcd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	ppath "path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula"
	"github.com/heetch/regula/errors"
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

func TestGet(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	t.Run("Root", func(t *testing.T) {
		path := "p/a/t/h"
		rs1, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))
		sig := regula.NewSignature(rs1)
		createRuleset(t, s, path, rs1)

		entry1, err := s.Get(context.Background(), path, "")
		require.NoError(t, err)
		require.Equal(t, path, entry1.Path)
		require.Equal(t, rs1, entry1.Ruleset)
		require.Equal(t, sig, entry1.Signature)
		require.NotEmpty(t, entry1.Version)
		require.Len(t, entry1.Versions, 1)

		// we are waiting 1 second here to avoid creating the new version in the same second as the previous one
		// ksuid gives a sorting with a one second precision
		time.Sleep(time.Second)
		rs2, _ := regula.NewBoolRuleset(rule.New(rule.Eq(rule.StringValue("foo"), rule.StringValue("foo")), rule.BoolValue(true)))
		createRuleset(t, s, path, rs2)

		// by default, it should return the latest version
		entry2, err := s.Get(context.Background(), path, "")
		require.NoError(t, err)
		require.Equal(t, path, entry2.Path)
		require.Equal(t, rs2, entry2.Ruleset)
		require.Equal(t, sig, entry2.Signature)
		require.NotEmpty(t, entry2.Version)
		require.Len(t, entry2.Versions, 2)

		// get a specific version
		entry3, err := s.Get(context.Background(), path, entry1.Version)
		require.NoError(t, err)
		require.Equal(t, entry1.Path, entry3.Path)
		require.Equal(t, entry1.Ruleset, entry3.Ruleset)
		require.Equal(t, entry1.Signature, entry3.Signature)
		require.Equal(t, entry1.Version, entry3.Version)
		require.Len(t, entry3.Versions, 2)
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := s.Get(context.Background(), "doesntexist", "")
		require.Equal(t, err, store.ErrNotFound)
	})
}

// List returns all rulesets entries or not depending on the query string.
func TestList(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	rs, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))

	// Root tests the basic behaviour without prefix.
	t.Run("Root", func(t *testing.T) {
		prefix := "list/root/"

		createRuleset(t, s, prefix+"c", rs)
		createRuleset(t, s, prefix+"a", rs)
		createRuleset(t, s, prefix+"a/1", rs)
		createRuleset(t, s, prefix+"b", rs)
		createRuleset(t, s, prefix+"a", rs)

		paths := []string{prefix + "a", prefix + "a/1", prefix + "b", prefix + "c"}

		entries, err := s.List(context.Background(), prefix+"", &store.ListOptions{})
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})

	// Assert that only latest version for each ruleset is returned by default.
	t.Run("Last version only", func(t *testing.T) {
		prefix := "list/last/version/"
		rs1, _ := regula.NewBoolRuleset(rule.New(rule.Eq(rule.BoolValue(true), rule.BoolValue(true)), rule.BoolValue(true)))
		rs2, _ := regula.NewBoolRuleset(rule.New(rule.Eq(rule.StringValue("true"), rule.StringValue("true")), rule.BoolValue(true)))

		createRuleset(t, s, prefix+"a", rs)
		createRuleset(t, s, prefix+"a/1", rs)
		createRuleset(t, s, prefix+"a", rs1)
		createRuleset(t, s, prefix+"a", rs2)

		entries, err := s.List(context.Background(), prefix+"", &store.ListOptions{})
		require.NoError(t, err)
		require.Len(t, entries.Entries, 2)
		a := entries.Entries[0]
		require.Equal(t, rs2, a.Ruleset)
		require.NotEmpty(t, entries.Revision)
	})

	// Assert that all versions are returned when passing the AllVersions option.
	t.Run("All versions", func(t *testing.T) {
		prefix := "list/all/version/"
		rs1, _ := regula.NewBoolRuleset(rule.New(rule.Eq(rule.BoolValue(true), rule.BoolValue(true)), rule.BoolValue(true)))
		rs2, _ := regula.NewBoolRuleset(rule.New(rule.Eq(rule.StringValue("true"), rule.StringValue("true")), rule.BoolValue(true)))

		createRuleset(t, s, prefix+"a", rs)
		time.Sleep(time.Second)
		createRuleset(t, s, prefix+"a", rs1)
		time.Sleep(time.Second)
		createRuleset(t, s, prefix+"a", rs2)
		createRuleset(t, s, prefix+"a/1", rs)

		paths := []string{prefix + "a", prefix + "a", prefix + "a", prefix + "a/1"}

		entries, err := s.List(context.Background(), prefix+"", &store.ListOptions{AllVersions: true})
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)

		// Assert that pagination is working well.
		opt := store.ListOptions{
			AllVersions: true,
			Limit:       2,
		}
		entries, err = s.List(context.Background(), prefix+"", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, opt.Limit)
		require.Equal(t, prefix+"a", entries.Entries[0].Path)
		require.Equal(t, rs, entries.Entries[0].Ruleset)
		require.Equal(t, prefix+"a", entries.Entries[1].Path)
		require.Equal(t, rs1, entries.Entries[1].Ruleset)
		require.NotEmpty(t, entries.Revision)

		opt.ContinueToken = entries.Continue
		entries, err = s.List(context.Background(), prefix+"", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, opt.Limit)
		require.Equal(t, prefix+"a", entries.Entries[0].Path)
		require.Equal(t, rs2, entries.Entries[0].Ruleset)
		require.Equal(t, prefix+"a/1", entries.Entries[1].Path)
		require.Equal(t, rs, entries.Entries[1].Ruleset)
		require.NotEmpty(t, entries.Revision)

		t.Run("NotFound", func(t *testing.T) {
			_, err = s.List(context.Background(), prefix+"doesntexist", &store.ListOptions{AllVersions: true})
			require.Equal(t, err, store.ErrNotFound)
		})

	})

	// Prefix tests List with a given prefix.
	t.Run("Prefix", func(t *testing.T) {
		prefix := "list/prefix/"

		createRuleset(t, s, prefix+"x", rs)
		createRuleset(t, s, prefix+"xx", rs)
		createRuleset(t, s, prefix+"x/1", rs)
		createRuleset(t, s, prefix+"x/2", rs)

		paths := []string{prefix + "x", prefix + "x/1", prefix + "x/2", prefix + "xx"}

		entries, err := s.List(context.Background(), prefix+"x", &store.ListOptions{})
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})

	// NotFound tests List with a prefix which doesn't exist.
	t.Run("NotFound", func(t *testing.T) {
		_, err := s.List(context.Background(), "doesntexist", &store.ListOptions{})
		require.Equal(t, err, store.ErrNotFound)
	})

	// Paging tests List with pagination.
	t.Run("Paging", func(t *testing.T) {
		prefix := "list/paging/"

		createRuleset(t, s, prefix+"y", rs)
		createRuleset(t, s, prefix+"yy", rs)
		createRuleset(t, s, prefix+"y/1", rs)
		createRuleset(t, s, prefix+"y/2", rs)
		createRuleset(t, s, prefix+"y/3", rs)

		opt := store.ListOptions{Limit: 2}
		entries, err := s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, 2)
		require.Equal(t, prefix+"y", entries.Entries[0].Path)
		require.Equal(t, prefix+"y/1", entries.Entries[1].Path)
		require.NotEmpty(t, entries.Continue)

		opt.ContinueToken = entries.Continue
		token := entries.Continue
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, 2)
		require.Equal(t, prefix+"y/2", entries.Entries[0].Path)
		require.Equal(t, prefix+"y/3", entries.Entries[1].Path)
		require.NotEmpty(t, entries.Continue)

		opt.ContinueToken = entries.Continue
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, 1)
		require.Equal(t, prefix+"yy", entries.Entries[0].Path)
		require.Empty(t, entries.Continue)

		opt.Limit = 3
		opt.ContinueToken = token
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, 3)
		require.Equal(t, prefix+"y/2", entries.Entries[0].Path)
		require.Equal(t, prefix+"y/3", entries.Entries[1].Path)
		require.Equal(t, prefix+"yy", entries.Entries[2].Path)
		require.Empty(t, entries.Continue)

		opt.ContinueToken = "some token"
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.Equal(t, store.ErrInvalidContinueToken, err)

		opt.Limit = -10
		opt.ContinueToken = ""
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, 5)
	})
}

// List returns all rulesets paths because the pathsOnly parameter is set to true.
// It returns all the entries or just a subset depending on the query string.
func TestListPaths(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	rs, _ := regula.NewBoolRuleset(rule.New(rule.True(), rule.BoolValue(true)))

	// Root is the basic behaviour without prefix with pathsOnly parameter set to true.
	t.Run("Root", func(t *testing.T) {
		prefix := "list/paths/root/"

		createRuleset(t, s, prefix+"a", rs)
		createRuleset(t, s, prefix+"b", rs)
		createRuleset(t, s, prefix+"a/1", rs)
		createRuleset(t, s, prefix+"c", rs)
		createRuleset(t, s, prefix+"a", rs)
		createRuleset(t, s, prefix+"a/1", rs)
		createRuleset(t, s, prefix+"a/2", rs)
		createRuleset(t, s, prefix+"d", rs)

		paths := []string{prefix + "a", prefix + "a/1", prefix + "a/2", prefix + "b", prefix + "c", prefix + "d"}

		opt := store.ListOptions{PathsOnly: true}
		entries, err := s.List(context.Background(), prefix+"", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Ruleset)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)
	})

	// Prefix tests List with a given prefix with pathsOnly parameter set to true.
	t.Run("Prefix", func(t *testing.T) {
		prefix := "list/paths/prefix/"

		createRuleset(t, s, prefix+"x", rs)
		createRuleset(t, s, prefix+"xx", rs)
		createRuleset(t, s, prefix+"x/1", rs)
		createRuleset(t, s, prefix+"xy", rs)
		createRuleset(t, s, prefix+"xy/ab", rs)
		createRuleset(t, s, prefix+"xyz", rs)

		paths := []string{prefix + "xy", prefix + "xy/ab", prefix + "xyz"}

		opt := store.ListOptions{PathsOnly: true}
		entries, err := s.List(context.Background(), prefix+"xy", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Ruleset)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)
	})

	// NotFound tests List with a prefix which doesn't exist with pathsOnly parameter set to true.
	t.Run("NotFound", func(t *testing.T) {
		opt := store.ListOptions{PathsOnly: true}
		_, err := s.List(context.Background(), "doesntexist", &opt)
		require.Equal(t, err, store.ErrNotFound)
	})

	// Paging tests List with pagination with pathsOnly parameter set to true.
	t.Run("Paging", func(t *testing.T) {
		prefix := "list/paths/paging/"

		createRuleset(t, s, prefix+"foo", rs)
		createRuleset(t, s, prefix+"foo/bar", rs)
		createRuleset(t, s, prefix+"foo/bar/baz", rs)
		createRuleset(t, s, prefix+"foo/bar", rs)
		createRuleset(t, s, prefix+"foo/babar", rs)
		createRuleset(t, s, prefix+"foo", rs)

		opt := store.ListOptions{Limit: 2, PathsOnly: true}
		entries, err := s.List(context.Background(), prefix+"f", &opt)
		require.NoError(t, err)
		paths := []string{prefix + "foo", prefix + "foo/babar"}
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Ruleset)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.NotEmpty(t, entries.Continue)

		opt.ContinueToken = entries.Continue
		entries, err = s.List(context.Background(), prefix+"f", &opt)
		require.NoError(t, err)
		paths = []string{prefix + "foo/bar", prefix + "foo/bar/baz"}
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Ruleset)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)

		opt.ContinueToken = "bad token"
		_, err = s.List(context.Background(), prefix+"f", &opt)
		require.Equal(t, store.ErrInvalidContinueToken, err)

		opt.Limit = -10
		opt.ContinueToken = ""
		entries, err = s.List(context.Background(), prefix+"f", &opt)
		require.NoError(t, err)
		paths = []string{prefix + "foo", prefix + "foo/babar", prefix + "foo/bar", prefix + "foo/bar/baz"}
		require.Len(t, entries.Entries, len(paths))
		for i, e := range entries.Entries {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Ruleset)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)
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
		require.Equal(t, entry.Version, strings.TrimPrefix(string(resp.Kvs[0].Key), ppath.Join(s.Namespace, "rulesets", "entries", "a")+"!"))

		// verify checksum creation
		resp, err = s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "checksums", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		// verify latest pointer creation
		resp, err = s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "latest", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		// verify versions list creation
		resp, err = s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "versions", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		var versions []string
		err = json.Unmarshal(resp.Kvs[0].Value, &versions)
		require.NoError(t, err)
		require.Len(t, versions, 1)
		require.EqualValues(t, entry.Version, versions[0])

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

		// verify versions list append
		resp, err = s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "versions", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		err = json.Unmarshal(resp.Kvs[0].Value, &versions)
		require.NoError(t, err)
		require.Len(t, versions, 2)
		require.EqualValues(t, entry.Version, versions[0])
		require.EqualValues(t, entry2.Version, versions[1])
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
		require.Equal(t, errors.ErrRulesetNotFound, err)
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
		require.Equal(t, errors.ErrRulesetNotFound, err)
	})

	t.Run("BadVersion", func(t *testing.T) {
		_, err := s.EvalVersion(context.Background(), "a", "someversion", regula.Params{
			"id": "123",
		})
		require.Equal(t, errors.ErrRulesetNotFound, err)
	})
}
