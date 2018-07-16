package client_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heetch/regula/rule"

	"github.com/heetch/regula/api"
	"github.com/heetch/regula/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleListRulesets() {
	c, err := client.NewClient("http://127.0.0.1:5331")
	if err != nil {
		log.Fatal(err)
	}

	list, err := c.ListRulesets(context.Background(), "prefix")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range list {
		e.Ruleset.Eval(nil)
	}
}

func ExampleEvalRuleset() {
	c, err := client.NewClient("http://127.0.0.1:5331")
	if err != nil {
		log.Fatal(err)
	}

	p := map[string]string{
		"foo": "bar",
		"baz": "42",
	}

	resp, err := c.EvalRuleset(context.Background(), "path/to/ruleset", p)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Data)
	fmt.Println(resp.Type)
}

func TestClient(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "some err"}`)
		}))
		defer ts.Close()

		cli, err := client.NewClient(ts.URL)
		require.NoError(t, err)

		_, err = cli.ListRulesets(context.Background(), "")
		require.EqualError(t, err, "some err")
	})

	t.Run("ListRulesets", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Contains(t, r.URL.Query(), "list")
			assert.Equal(t, "/rulesets/prefix", r.URL.Path)
			fmt.Fprintf(w, `[{"path": "a"}]`)
		}))
		defer ts.Close()

		cli, err := client.NewClient(ts.URL)
		require.NoError(t, err)

		rs, err := cli.ListRulesets(context.Background(), "prefix")
		require.NoError(t, err)
		require.Len(t, rs, 1)
	})

	t.Run("EvalRuleset", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Contains(t, r.URL.Query(), "eval")
			assert.Contains(t, r.URL.Query(), "foo")
			assert.Equal(t, "/rulesets/path/to/ruleset", r.URL.Path)
			fmt.Fprintf(w, `{"data": "baz", "type": "string"}]`)
		}))
		defer ts.Close()

		cli, err := client.NewClient(ts.URL)
		require.NoError(t, err)

		p := map[string]string{
			"foo": "bar",
		}

		exp := api.Value{
			Data: "baz",
			Type: "string",
		}

		resp, err := cli.EvalRuleset(context.Background(), "path/to/ruleset", p)
		require.NoError(t, err)
		require.Equal(t, &exp, resp)
	})

	t.Run("PutRuleset", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "/rulesets/a", r.URL.Path)
			fmt.Fprintf(w, `{"path": "a", "version": "v"}`)
		}))
		defer ts.Close()

		cli, err := client.NewClient(ts.URL)
		require.NoError(t, err)

		rs, err := rule.NewInt64Ruleset(rule.New(rule.True(), rule.ReturnsInt64(1)))
		require.NoError(t, err)

		ars, err := cli.PutRuleset(context.Background(), "a", rs)
		require.NoError(t, err)
		require.Equal(t, "a", ars.Path)
		require.Equal(t, "v", ars.Version)
	})
}
