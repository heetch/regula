package consul

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/heetch/rules-engine/rule"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	consulAddr := "127.0.0.1:8500"
	keyPrefix, insert, teardown := consulHelper(t, consulAddr)
	defer teardown()

	t.Run("/a/b/c", func(t *testing.T) {
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
		raw, err := json.Marshal(rule.Ruleset{r1, rd})
		require.NoError(t, err)

		insert("a/b/c", raw)

		s, err := NewStore("127.0.0.1:8500", keyPrefix)
		require.NoError(t, err)

		rs, err := s.Get("a/b/c")
		require.NoError(t, err)

		require.Len(t, rs, 2)

		res, err := rs.Eval(rule.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, "string", res.Type)
		require.Equal(t, "matched r1", res.Value)
	})
}

func consulHelper(t *testing.T, consulAddr string) (string, func(string, []byte), func()) {
	keyPrefix := fmt.Sprintf("testing/%d/", time.Now().Unix())

	conf := &api.Config{
		Scheme:  "http",
		Address: consulAddr,
	}
	client, err := api.NewClient(conf)

	if err != nil {
		t.Fatal(err)
	}

	insert := func(key string, value []byte) {
		_, err := client.KV().Put(&api.KVPair{Key: keyPrefix + key, Value: value}, nil)
		require.NoError(t, err)
		return
	}

	teardown := func() {
		_, err := client.KV().DeleteTree(keyPrefix, nil)
		require.NoError(t, err)
		return
	}

	return keyPrefix, insert, teardown
}
