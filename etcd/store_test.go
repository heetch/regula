package etcd_test

import (
	"context"
	"encoding/json"
	"log"
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

func etcdHelper(t *testing.T) (*clientv3.Client, func()) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	return cli, func() {
		cli.Delete(context.Background(), "", clientv3.WithPrefix())
		cli.Close()
	}
}

func TestEtcdStore(t *testing.T) {
	cli, cleanup := etcdHelper(t)
	defer cleanup()

	prefix := "keys"

	r1 := rule.New(
		rule.Eq(
			rule.ValueStr("bar"),
			rule.ParamStr("foo"),
		),
		rule.ReturnsStr("matched r1"),
	)

	rd := rule.New(
		rule.True(),
		rule.ReturnsStr("matched default"),
	)
	rs, err := rule.NewStringRuleset(r1, rd)
	require.NoError(t, err)

	raw, err := json.Marshal(rs)
	require.NoError(t, err)

	_, err = cli.Put(context.Background(), path.Join(prefix, "a/b/c"), string(raw))
	require.NoError(t, err)

	store, err := etcd.NewStore(cli, prefix)
	require.NoError(t, err)

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
				require.Equal(t, "matched r1", res.Value)
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

func Example() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{":2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	store, err := etcd.NewStore(cli, "prefix")
	if err != nil {
		log.Fatal(err)
	}

	_, err = store.Get("some-key")
	if err != nil {
		log.Fatal(err)
	}
}
