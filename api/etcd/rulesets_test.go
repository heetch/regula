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
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/api/etcd"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

var (
	_ api.RulesetService = new(etcd.RulesetService)
	_ regula.Evaluator   = new(etcd.RulesetService)
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

func createRuleset(t *testing.T, s *etcd.RulesetService, path string, rules ...*rule.Rule) *regula.Ruleset {
	e, err := s.Put(context.Background(), path, rules)
	if err != nil && err != api.ErrRulesetNotModified {
		require.NoError(t, err)
	}
	return e
}

func createBoolRuleset(t *testing.T, s *etcd.RulesetService, path string, rules ...*rule.Rule) *regula.Ruleset {
	err := s.Create(context.Background(), path, &regula.Signature{ReturnType: "bool"})
	require.False(t, err != nil && err != api.ErrAlreadyExists)
	return createRuleset(t, s, path, rules...)
}

func TestGet(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	path := "p/a/t/h"
	sig := &regula.Signature{ReturnType: "bool", Params: make(map[string]string)}

	t.Run("Root", func(t *testing.T) {
		rs1 := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}
		createBoolRuleset(t, s, path, rs1...)

		entry1, err := s.Get(context.Background(), path, "")
		require.NoError(t, err)
		require.Equal(t, path, entry1.Path)
		require.Equal(t, rs1, entry1.Rules)
		require.Equal(t, sig, entry1.Signature)
		require.NotEmpty(t, entry1.Version)
		require.Len(t, entry1.Versions, 1)

		// we are waiting 1 second here to avoid creating the new version in the same second as the previous one
		// ksuid gives a sorting with a one second precision
		time.Sleep(time.Second)
		rs2 := []*rule.Rule{rule.New(rule.Eq(rule.StringValue("foo"), rule.StringValue("foo")), rule.BoolValue(true))}
		createBoolRuleset(t, s, path, rs2...)

		// by default, it should return the latest version
		entry2, err := s.Get(context.Background(), path, "")
		require.NoError(t, err)
		require.Equal(t, path, entry2.Path)
		require.Equal(t, rs2, entry2.Rules)
		require.Equal(t, sig, entry2.Signature)
		require.NotEmpty(t, entry2.Version)
		require.Len(t, entry2.Versions, 2)

		// get a specific version
		entry3, err := s.Get(context.Background(), path, entry1.Version)
		require.NoError(t, err)
		require.Equal(t, entry1.Path, entry3.Path)
		require.Equal(t, entry1.Rules, entry3.Rules)
		require.Equal(t, entry1.Signature, entry3.Signature)
		require.Equal(t, entry1.Version, entry3.Version)
		require.Len(t, entry3.Versions, 2)
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := s.Get(context.Background(), "doesntexist", "")
		require.Equal(t, err, api.ErrRulesetNotFound)
	})
}

// List returns all rulesets entries or not depending on the query string.
func TestList(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	rs := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}

	// Root tests the basic behaviour without prefix.
	t.Run("Root", func(t *testing.T) {
		prefix := "list/root/"

		createBoolRuleset(t, s, prefix+"c", rs...)
		createBoolRuleset(t, s, prefix+"a", rs...)
		createBoolRuleset(t, s, prefix+"a/1", rs...)
		createBoolRuleset(t, s, prefix+"b", rs...)
		createBoolRuleset(t, s, prefix+"a", rs...)

		paths := []string{prefix + "a", prefix + "a/1", prefix + "b", prefix + "c"}

		entries, err := s.List(context.Background(), prefix+"", &api.ListOptions{})
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})

	// Assert that only latest version for each ruleset is returned by default.
	t.Run("Last version only", func(t *testing.T) {
		prefix := "list/last/version/"
		rs1 := []*rule.Rule{rule.New(rule.Eq(rule.BoolValue(true), rule.BoolValue(true)), rule.BoolValue(true))}
		rs2 := []*rule.Rule{rule.New(rule.Eq(rule.StringValue("true"), rule.StringValue("true")), rule.BoolValue(true))}

		createBoolRuleset(t, s, prefix+"a", rs...)
		createBoolRuleset(t, s, prefix+"a/1", rs...)
		createBoolRuleset(t, s, prefix+"a", rs1...)
		createBoolRuleset(t, s, prefix+"a", rs2...)

		entries, err := s.List(context.Background(), prefix+"", &api.ListOptions{})
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, 2)
		a := entries.Rulesets[0]
		require.Equal(t, rs2, a.Rules)
		require.NotEmpty(t, entries.Revision)
	})

	// Assert that all versions are returned when passing the AllVersions option.
	t.Run("All versions", func(t *testing.T) {
		prefix := "list/all/version/"
		rs1 := []*rule.Rule{rule.New(rule.Eq(rule.BoolValue(true), rule.BoolValue(true)), rule.BoolValue(true))}
		rs2 := []*rule.Rule{rule.New(rule.Eq(rule.StringValue("true"), rule.StringValue("true")), rule.BoolValue(true))}

		createBoolRuleset(t, s, prefix+"a", rs...)
		time.Sleep(time.Second)
		createBoolRuleset(t, s, prefix+"a", rs1...)
		time.Sleep(time.Second)
		createBoolRuleset(t, s, prefix+"a", rs2...)
		createBoolRuleset(t, s, prefix+"a/1", rs...)

		paths := []string{prefix + "a", prefix + "a", prefix + "a", prefix + "a/1"}

		entries, err := s.List(context.Background(), prefix+"", &api.ListOptions{AllVersions: true})
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)

		// Assert that pagination is working well.
		opt := api.ListOptions{
			AllVersions: true,
			Limit:       2,
		}
		entries, err = s.List(context.Background(), prefix+"", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, opt.Limit)
		require.Equal(t, prefix+"a", entries.Rulesets[0].Path)
		require.Equal(t, rs, entries.Rulesets[0].Rules)
		require.Equal(t, prefix+"a", entries.Rulesets[1].Path)
		require.Equal(t, rs1, entries.Rulesets[1].Rules)
		require.NotEmpty(t, entries.Revision)

		opt.ContinueToken = entries.Continue
		entries, err = s.List(context.Background(), prefix+"", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, opt.Limit)
		require.Equal(t, prefix+"a", entries.Rulesets[0].Path)
		require.Equal(t, rs2, entries.Rulesets[0].Rules)
		require.Equal(t, prefix+"a/1", entries.Rulesets[1].Path)
		require.Equal(t, rs, entries.Rulesets[1].Rules)
		require.NotEmpty(t, entries.Revision)

		t.Run("NotFound", func(t *testing.T) {
			_, err = s.List(context.Background(), prefix+"doesntexist", &api.ListOptions{AllVersions: true})
			require.Equal(t, err, api.ErrRulesetNotFound)
		})

	})

	// Prefix tests List with a given prefix.
	t.Run("Prefix", func(t *testing.T) {
		prefix := "list/prefix/"

		createBoolRuleset(t, s, prefix+"x", rs...)
		createBoolRuleset(t, s, prefix+"xx", rs...)
		createBoolRuleset(t, s, prefix+"x/1", rs...)
		createBoolRuleset(t, s, prefix+"x/2", rs...)

		paths := []string{prefix + "x", prefix + "x/1", prefix + "x/2", prefix + "xx"}

		entries, err := s.List(context.Background(), prefix+"x", &api.ListOptions{})
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
		}
		require.NotEmpty(t, entries.Revision)
	})

	// NotFound tests List with a prefix which doesn't exist.
	t.Run("NotFound", func(t *testing.T) {
		_, err := s.List(context.Background(), "doesntexist", &api.ListOptions{})
		require.Equal(t, err, api.ErrRulesetNotFound)
	})

	// Paging tests List with pagination.
	t.Run("Paging", func(t *testing.T) {
		prefix := "list/paging/"

		createBoolRuleset(t, s, prefix+"y", rs...)
		createBoolRuleset(t, s, prefix+"yy", rs...)
		createBoolRuleset(t, s, prefix+"y/1", rs...)
		createBoolRuleset(t, s, prefix+"y/2", rs...)
		createBoolRuleset(t, s, prefix+"y/3", rs...)

		opt := api.ListOptions{Limit: 2}
		entries, err := s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, 2)
		require.Equal(t, prefix+"y", entries.Rulesets[0].Path)
		require.Equal(t, prefix+"y/1", entries.Rulesets[1].Path)
		require.NotEmpty(t, entries.Continue)

		opt.ContinueToken = entries.Continue
		token := entries.Continue
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, 2)
		require.Equal(t, prefix+"y/2", entries.Rulesets[0].Path)
		require.Equal(t, prefix+"y/3", entries.Rulesets[1].Path)
		require.NotEmpty(t, entries.Continue)

		opt.ContinueToken = entries.Continue
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, 1)
		require.Equal(t, prefix+"yy", entries.Rulesets[0].Path)
		require.Empty(t, entries.Continue)

		opt.Limit = 3
		opt.ContinueToken = token
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, 3)
		require.Equal(t, prefix+"y/2", entries.Rulesets[0].Path)
		require.Equal(t, prefix+"y/3", entries.Rulesets[1].Path)
		require.Equal(t, prefix+"yy", entries.Rulesets[2].Path)
		require.Empty(t, entries.Continue)

		opt.ContinueToken = "some token"
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.Equal(t, api.ErrInvalidContinueToken, err)

		opt.Limit = -10
		opt.ContinueToken = ""
		entries, err = s.List(context.Background(), prefix+"y", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, 5)
	})
}

// List returns all rulesets paths because the pathsOnly parameter is set to true.
// It returns all the entries or just a subset depending on the query string.
func TestListPaths(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	rs := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}

	// Root is the basic behaviour without prefix with pathsOnly parameter set to true.
	t.Run("Root", func(t *testing.T) {
		prefix := "list/paths/root/"

		createBoolRuleset(t, s, prefix+"a", rs...)
		createBoolRuleset(t, s, prefix+"b", rs...)
		createBoolRuleset(t, s, prefix+"a/1", rs...)
		createBoolRuleset(t, s, prefix+"c", rs...)
		createBoolRuleset(t, s, prefix+"a", rs...)
		createBoolRuleset(t, s, prefix+"a/1", rs...)
		createBoolRuleset(t, s, prefix+"a/2", rs...)
		createBoolRuleset(t, s, prefix+"d", rs...)

		paths := []string{prefix + "a", prefix + "a/1", prefix + "a/2", prefix + "b", prefix + "c", prefix + "d"}

		opt := api.ListOptions{PathsOnly: true}
		entries, err := s.List(context.Background(), prefix+"", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Rules)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)
	})

	// Prefix tests List with a given prefix with pathsOnly parameter set to true.
	t.Run("Prefix", func(t *testing.T) {
		prefix := "list/paths/prefix/"

		createBoolRuleset(t, s, prefix+"x", rs...)
		createBoolRuleset(t, s, prefix+"xx", rs...)
		createBoolRuleset(t, s, prefix+"x/1", rs...)
		createBoolRuleset(t, s, prefix+"xy", rs...)
		createBoolRuleset(t, s, prefix+"xy/ab", rs...)
		createBoolRuleset(t, s, prefix+"xyz", rs...)

		paths := []string{prefix + "xy", prefix + "xy/ab", prefix + "xyz"}

		opt := api.ListOptions{PathsOnly: true}
		entries, err := s.List(context.Background(), prefix+"xy", &opt)
		require.NoError(t, err)
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Rules)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)
	})

	// NotFound tests List with a prefix which doesn't exist with pathsOnly parameter set to true.
	t.Run("NotFound", func(t *testing.T) {
		opt := api.ListOptions{PathsOnly: true}
		_, err := s.List(context.Background(), "doesntexist", &opt)
		require.Equal(t, err, api.ErrRulesetNotFound)
	})

	// Paging tests List with pagination with pathsOnly parameter set to true.
	t.Run("Paging", func(t *testing.T) {
		prefix := "list/paths/paging/"

		createBoolRuleset(t, s, prefix+"foo", rs...)
		createBoolRuleset(t, s, prefix+"foo/bar", rs...)
		createBoolRuleset(t, s, prefix+"foo/bar/baz", rs...)
		createBoolRuleset(t, s, prefix+"foo/bar", rs...)
		createBoolRuleset(t, s, prefix+"foo/babar", rs...)
		createBoolRuleset(t, s, prefix+"foo", rs...)

		opt := api.ListOptions{Limit: 2, PathsOnly: true}
		entries, err := s.List(context.Background(), prefix+"f", &opt)
		require.NoError(t, err)
		paths := []string{prefix + "foo", prefix + "foo/babar"}
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Rules)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.NotEmpty(t, entries.Continue)

		opt.ContinueToken = entries.Continue
		entries, err = s.List(context.Background(), prefix+"f", &opt)
		require.NoError(t, err)
		paths = []string{prefix + "foo/bar", prefix + "foo/bar/baz"}
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Rules)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)

		opt.ContinueToken = "bad token"
		_, err = s.List(context.Background(), prefix+"f", &opt)
		require.Equal(t, api.ErrInvalidContinueToken, err)

		opt.Limit = -10
		opt.ContinueToken = ""
		entries, err = s.List(context.Background(), prefix+"f", &opt)
		require.NoError(t, err)
		paths = []string{prefix + "foo", prefix + "foo/babar", prefix + "foo/bar", prefix + "foo/bar/baz"}
		require.Len(t, entries.Rulesets, len(paths))
		for i, e := range entries.Rulesets {
			require.Equal(t, paths[i], e.Path)
			require.Zero(t, e.Rules)
			require.Zero(t, e.Version)
		}
		require.NotEmpty(t, entries.Revision)
		require.Zero(t, entries.Continue)
	})
}

func TestPut(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	path := "a"
	sig := &regula.Signature{ReturnType: "bool"}
	require.NoError(t, s.Create(context.Background(), path, sig))

	t.Run("OK", func(t *testing.T) {
		path := "a"
		rs := []*rule.Rule{rule.New(rule.True(), rule.BoolValue(true))}

		entry, err := s.Put(context.Background(), path, rs)
		require.NoError(t, err)
		require.Equal(t, path, entry.Path)
		require.NotEmpty(t, entry.Version)
		require.Equal(t, rs, entry.Rules)

		// verify ruleset creation
		resp, err := s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "rules", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)
		// verify if the path contains the right ruleset version
		require.Equal(t, entry.Version, strings.TrimPrefix(string(resp.Kvs[0].Key), ppath.Join(s.Namespace, "rulesets", "rules", "a")+"!"))

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

		var v pb.Versions
		err = proto.Unmarshal(resp.Kvs[0].Value, &v)
		require.NoError(t, err)
		require.Len(t, v.Versions, 1)
		require.EqualValues(t, entry.Version, v.Versions[0])

		// create new version with same ruleset
		entry2, err := s.Put(context.Background(), path, rs)
		require.Equal(t, api.ErrRulesetNotModified, err)
		require.Equal(t, entry, entry2)

		// create new version with different ruleset
		rs = []*rule.Rule{rule.New(rule.True(), rule.BoolValue(false))}

		entry2, err = s.Put(context.Background(), path, rs)
		require.NoError(t, err)
		require.NotEqual(t, entry.Version, entry2.Version)

		// verify versions list append
		resp, err = s.Client.Get(context.Background(), ppath.Join(s.Namespace, "rulesets", "versions", path), clientv3.WithPrefix())
		require.NoError(t, err)
		require.EqualValues(t, 1, resp.Count)

		err = proto.Unmarshal(resp.Kvs[0].Value, &v)
		require.NoError(t, err)
		require.Len(t, v.Versions, 2)
		require.EqualValues(t, entry.Version, v.Versions[0])
		require.EqualValues(t, entry2.Version, v.Versions[1])
	})

	t.Run("Signatures", func(t *testing.T) {
		path := "b"
		require.NoError(t, s.Create(context.Background(), path, &regula.Signature{ReturnType: "bool", Params: map[string]string{"a": "string", "b": "bool", "c": "int64"}}))

		rs1 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
				),
				rule.BoolValue(true),
			),
		}

		_, err := s.Put(context.Background(), path, rs1)
		require.NoError(t, err)

		// same params, different return type
		rs2 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
				),
				rule.StringValue("true"),
			),
		}

		_, err = s.Put(context.Background(), path, rs2)
		require.True(t, api.IsValidationError(err))

		// adding new params
		rs3 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.BoolParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		}

		_, err = s.Put(context.Background(), path, rs3)
		require.True(t, api.IsValidationError(err))

		// changing param types
		rs4 := []*rule.Rule{
			rule.New(
				rule.Eq(
					rule.StringParam("a"),
					rule.StringParam("b"),
					rule.Int64Param("c"),
					rule.BoolParam("d"),
				),
				rule.BoolValue(true),
			),
		}

		_, err = s.Put(context.Background(), path, rs4)
		require.True(t, api.IsValidationError(err))

		// adding new rule with different param types
		rs5 := []*rule.Rule{
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
		}

		_, err = s.Put(context.Background(), path, rs5)
		require.True(t, api.IsValidationError(err))

		// adding new rule with correct param types but less
		rs6 := []*rule.Rule{
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
		}

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

		r := rule.New(rule.True(), rule.BoolValue(true))

		createBoolRuleset(t, s, "aa", r)
		createBoolRuleset(t, s, "ab", r)
		createBoolRuleset(t, s, "a/1", r)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events, err := s.Watch(ctx, "a", "")
	require.NoError(t, err)
	require.Len(t, events.Events, 1)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, "aa", events.Events[0].Path)
	require.Equal(t, api.RulesetPutEvent, events.Events[0].Type)

	wg.Wait()

	events, err = s.Watch(ctx, "a", events.Revision)
	require.NoError(t, err)
	require.Len(t, events.Events, 2)
	require.NotEmpty(t, events.Revision)
	require.Equal(t, api.RulesetPutEvent, events.Events[0].Type)
	require.Equal(t, "ab", events.Events[0].Path)
	require.Equal(t, api.RulesetPutEvent, events.Events[1].Type)
	require.Equal(t, "a/1", events.Events[1].Path)
}

func TestEval(t *testing.T) {
	t.Parallel()

	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	sig := &regula.Signature{ReturnType: "bool", Params: map[string]string{"id": "string"}}
	require.NoError(t, s.Create(context.Background(), "a", sig))

	r := rule.New(
		rule.Eq(
			rule.StringParam("id"),
			rule.StringValue("123"),
		),
		rule.BoolValue(true),
	)

	entry := createRuleset(t, s, "a", r)

	t.Run("OK", func(t *testing.T) {
		res, err := s.Eval(context.Background(), "a", entry.Version, regula.Params{
			"id": "123",
		})
		require.NoError(t, err)
		require.Equal(t, entry.Version, res.Version)
		require.Equal(t, rule.BoolValue(true), res.Value)
	})

	t.Run("Latest", func(t *testing.T) {
		res, err := s.Eval(context.Background(), "a", "", regula.Params{
			"id": "123",
		})
		require.NoError(t, err)
		require.Equal(t, entry.Version, res.Version)
		require.Equal(t, rule.BoolValue(true), res.Value)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := s.Eval(context.Background(), "b", entry.Version, regula.Params{
			"id": "123",
		})
		require.Equal(t, errors.ErrRulesetNotFound, err)
	})

	t.Run("BadVersion", func(t *testing.T) {
		_, err := s.Eval(context.Background(), "a", "someversion", regula.Params{
			"id": "123",
		})
		require.Equal(t, errors.ErrRulesetNotFound, err)
	})
}
