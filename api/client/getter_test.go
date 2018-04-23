package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/api/client"
	"github.com/stretchr/testify/require"
)

func TestGetter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[{"name": "a", "ruleset": {"type": "string"}}]`)
	}))
	defer ts.Close()

	cli, err := client.NewClient(ts.URL)
	require.NoError(t, err)

	g, err := client.NewGetter(context.Background(), cli)
	require.NoError(t, err)

	rs, err := g.Get(nil, "a", nil)
	require.NoError(t, err)
	require.Equal(t, "string", rs.Type)

	rs, err = g.Get(nil, "b", nil)
	require.Equal(t, rules.ErrRulesetNotFound, err)
}
