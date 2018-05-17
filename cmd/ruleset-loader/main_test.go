package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/coreos/etcd/clientv3"

	"github.com/heetch/rules-engine/rule"
	"github.com/heetch/rules-engine/store/etcd"
	"github.com/stretchr/testify/require"
)

func TestLoadSnapshot(t *testing.T) {
	rs, err := rule.NewInt64Ruleset(
		rule.New(rule.True(), rule.ReturnsInt64(10)),
	)
	require.NoError(t, err)
	raw, err := json.Marshal(rs)
	require.NoError(t, err)

	snapshot := `{
		"/snapshot-tests/a/": ` + string(raw) + `,
		"/snapshot-tests/b": ` + string(raw) + `,
		"snapshot-tests/c/": ` + string(raw) + `,
		"snapshot-tests/d": ` + string(raw) + `
	}`

	err = loadSnapshot(strings.NewReader(snapshot))
	require.NoError(t, err)

	client, err := newClient()
	require.NoError(t, err)
	defer client.Close()
	defer client.Delete(context.Background(), "snapshot-tests/", clientv3.WithPrefix())

	store := etcd.Store{
		Client:    client,
		Namespace: "snapshot-tests",
	}

	entries, err := store.All(context.Background())
	require.NoError(t, err)

	require.Len(t, entries, 4)
	require.Equal(t, "snapshot-tests/a", entries[0].Name)
	require.Equal(t, "snapshot-tests/b", entries[1].Name)
	require.Equal(t, "snapshot-tests/c", entries[2].Name)
	require.Equal(t, "snapshot-tests/d", entries[3].Name)
}
