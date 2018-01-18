package etcd_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/etcd"
	"github.com/heetch/rules-engine/rule"
	"github.com/stretchr/testify/require"
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"localhost:2379"}
)

func etcdHelper(t *testing.T) (*clientv3.Client, string, func()) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	return cli, "rules-engine-tests", func() {
		cli.Delete(context.Background(), "rules-engine-tests", clientv3.WithPrefix())
		cli.Close()
	}
}

func TestEtcdStore(t *testing.T) {
	cli, prefix, cleanup := etcdHelper(t)
	defer cleanup()

	createRuleset(t, cli, prefix, "a/b/c")

	store, err := etcd.NewStore(cli, etcd.Options{Prefix: prefix})
	require.NoError(t, err)
	defer store.Close()

	t.Run("OK", func(t *testing.T) {
		tests := []string{
			"/a/b/c",
			"a/b/c",
			"a/b/c/",
			"/a/b/c/",
		}

		for _, test := range tests {
			t.Run(test, func(t *testing.T) {
				rs, err := store.Get(test)
				require.NoError(t, err)
				require.Len(t, rs.Rules, 2)

				res, err := rs.Eval(rule.Params{
					"foo": "bar",
				})
				require.NoError(t, err)
				require.Equal(t, "string", res.Type)
				require.Equal(t, "matched r1", res.Data)
			})
		}
	})

	t.Run("Ruleset not found", func(t *testing.T) {
		_, err := store.Get("unknown")
		require.Equal(t, rules.ErrRulesetNotFound, err)

		_, err = store.Get("")
		require.Equal(t, rules.ErrRulesetNotFound, err)
	})
}

func TestEtcdStoreWatcher(t *testing.T) {
	cli, prefix, cleanup := etcdHelper(t)
	defer cleanup()

	createRuleset(t, cli, prefix, "a")

	store, err := etcd.NewStore(cli, etcd.Options{Prefix: prefix})
	require.NoError(t, err)
	defer store.Close()

	createRuleset(t, cli, prefix, "b")

	var found bool
	for i := 0; i < 50; i++ {
		r, err := store.Get("b")
		if err == nil {
			found = true
			require.NotEmpty(t, r)
			break
		}

		if err != rules.ErrRulesetNotFound {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	if !found {
		t.Fatal("watcher not working properly, PUT not catched")
	}

	deleteRuleset(t, cli, prefix, "b")

	var deleted bool
	for i := 0; i < 50; i++ {
		_, err := store.Get("b")
		if err == rules.ErrRulesetNotFound {
			deleted = true
			break
		}

		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	if !deleted {
		t.Fatal("watcher not working properly, DELETE not catched")
	}
}

func createRuleset(t *testing.T, client *clientv3.Client, prefix, name string) {
	r1 := rule.New(
		rule.Eq(
			rule.StringValue("bar"),
			rule.StringParam("foo"),
		),
		rule.ReturnsString("matched r1"),
	)

	rd := rule.New(
		rule.True(),
		rule.ReturnsString("matched default"),
	)
	rs, err := rule.NewStringRuleset(r1, rd)
	require.NoError(t, err)

	raw, err := json.Marshal(rs)
	require.NoError(t, err)

	_, err = client.Put(context.Background(), path.Join(prefix, name), string(raw))
	require.NoError(t, err)
}

func deleteRuleset(t *testing.T, client *clientv3.Client, prefix, name string) {
	resp, err := client.Delete(context.Background(), path.Join(prefix, name))
	require.NoError(t, err)
	require.EqualValues(t, 1, resp.Deleted)
}

func Example() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{":2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	store, err := etcd.NewStore(cli, etcd.Options{
		Prefix: "prefix",
		Logger: log.New(os.Stdout, "[etcd] ", log.LstdFlags),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	_, err = store.Get("some-key")
	if err != nil {
		log.Fatal(err)
	}
}
