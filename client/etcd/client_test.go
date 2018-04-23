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
	"github.com/heetch/rules-engine/client"
	"github.com/heetch/rules-engine/client/etcd"
	"github.com/heetch/rules-engine/rule"
	"github.com/heetch/rules-engine/store"
	"github.com/stretchr/testify/require"
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"localhost:2379", "etcd:2379"}
)

func etcdHelper(t *testing.T, namespace string) (*clientv3.Client, func()) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	return cli, func() {
		cli.Delete(context.Background(), namespace, clientv3.WithPrefix())
		cli.Close()
	}
}

func TestEtcdClientGet(t *testing.T) {
	ns := "rules-engine-tests"
	cli, cleanup := etcdHelper(t, ns)
	defer cleanup()

	createRuleset(t, cli, ns, "a/b/c")

	st, err := etcd.NewClient(cli, etcd.Options{Namespace: ns})
	require.NoError(t, err)
	defer st.Close()

	t.Run("OK", func(t *testing.T) {
		tests := []string{
			"/a/b/c",
			"a/b/c",
			"a/b/c/",
			"/a/b/c/",
		}

		for _, test := range tests {
			t.Run(test, func(t *testing.T) {
				rs, err := st.Get(context.Background(), test)
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
		_, err := st.Get(context.Background(), "unknown")
		require.Equal(t, client.ErrRulesetNotFound, err)

		_, err = st.Get(context.Background(), "")
		require.Equal(t, client.ErrRulesetNotFound, err)
	})
}

func TestEtcdNamespacing(t *testing.T) {
	ns := "ns1"
	cli, cleanup := etcdHelper(t, ns)
	defer cleanup()

	createRuleset(t, cli, ns, "a")
	createRuleset(t, cli, ns+"2", "a")

	st, err := etcd.NewClient(cli, etcd.Options{Namespace: ns})
	require.NoError(t, err)
	defer st.Close()

	_, err = st.Get(context.Background(), "a")
	require.NoError(t, err)

	_, err = st.Get(context.Background(), "2/a")
	require.Equal(t, err, client.ErrRulesetNotFound)
}

func TestEtcdClientWatcher(t *testing.T) {
	ns := "rules-engine-tests"
	cli, cleanup := etcdHelper(t, ns)
	defer cleanup()

	createRuleset(t, cli, ns, "a")

	st, err := etcd.NewClient(cli, etcd.Options{Namespace: ns})
	require.NoError(t, err)
	defer st.Close()

	createRuleset(t, cli, ns, "b")

	var found bool
	for i := 0; i < 50; i++ {
		r, err := st.Get(context.Background(), "b")
		if err == nil {
			found = true
			require.NotEmpty(t, r)
			break
		}

		if err != client.ErrRulesetNotFound {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	if !found {
		t.Fatal("watcher not working properly, PUT not catched")
	}

	deleteRuleset(t, cli, ns, "b")

	var deleted bool
	for i := 0; i < 50; i++ {
		_, err := st.Get(context.Background(), "b")
		if err == client.ErrRulesetNotFound {
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

func createRuleset(t *testing.T, client *clientv3.Client, namespace, name string) {
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

	re := store.RulesetEntry{
		Name:    name,
		Ruleset: rs,
	}

	raw, err := json.Marshal(re)
	require.NoError(t, err)

	_, err = client.Put(context.Background(), path.Join(namespace, name), string(raw))
	require.NoError(t, err)
}

func deleteRuleset(t *testing.T, client *clientv3.Client, namespace, name string) {
	resp, err := client.Delete(context.Background(), path.Join(namespace, name))
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

	st, err := etcd.NewClient(cli, etcd.Options{
		Namespace: "namespace",
		Logger:    log.New(os.Stdout, "[etcd] ", log.LstdFlags),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	_, err = st.Get(context.Background(), "some-key")
	if err != nil {
		log.Fatal(err)
	}
}
