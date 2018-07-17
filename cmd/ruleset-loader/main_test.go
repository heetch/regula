package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heetch/regula/api/client"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSnapshot(t *testing.T) {
	var counter int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/rulesets/snapshot-tests/%c", 'a'+byte(counter)), r.URL.Path)
		counter++
		fmt.Fprintf(w, `{"path": "a", "version": "v"}]`)
	}))
	defer ts.Close()

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

	client, err := client.New(ts.URL)
	require.NoError(t, err)

	err = loadSnapshot(client, strings.NewReader(snapshot))
	require.NoError(t, err)
}
