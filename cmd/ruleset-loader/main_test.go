package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api/client"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSnapshot(t *testing.T) {
	var counter int

	paths := []string{"a", "b", "c", "d"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, paths, strings.TrimPrefix(r.URL.Path, "/rulesets/snapshot-tests/"))
		counter++
		fmt.Fprintf(w, `{"path": "a", "version": "v"}]`)
	}))
	defer ts.Close()

	rs, err := regula.NewInt64Ruleset(
		regula.NewRule(regula.True(), regula.Int64Value(10)),
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
	client.Logger = zerolog.New(ioutil.Discard)

	err = loadSnapshot(client, strings.NewReader(snapshot))
	require.NoError(t, err)
}
