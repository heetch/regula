package client_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/api/client"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleClient_ListRulesets() {
	c, err := client.New("http://127.0.0.1:5331")
	if err != nil {
		log.Fatal(err)
	}

	list, err := c.ListRulesets(context.Background(), "prefix")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range list.Rulesets {
		e.Ruleset.Eval(nil)
	}
}

func ExampleClient_EvalRuleset() {
	c, err := client.New("http://127.0.0.1:5331")
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

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		_, err = cli.ListRulesets(context.Background(), "")
		aerr := err.(*api.Error)
		require.Equal(t, "some err", aerr.Err)
	})

	t.Run("ListRulesets", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Contains(t, r.URL.Query(), "list")
			assert.Equal(t, "/rulesets/prefix", r.URL.Path)
			fmt.Fprintf(w, `{"revision": "rev", "rulesets": [{"path": "a"}]}`)
		}))
		defer ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		rs, err := cli.ListRulesets(context.Background(), "prefix")
		require.NoError(t, err)
		require.Len(t, rs.Rulesets, 1)
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

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

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

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		rs, err := regula.NewInt64Ruleset(regula.NewRule(regula.True(), regula.ReturnsInt64(1)))
		require.NoError(t, err)

		ars, err := cli.PutRuleset(context.Background(), "a", rs)
		require.NoError(t, err)
		require.Equal(t, "a", ars.Path)
		require.Equal(t, "v", ars.Version)
	})

	t.Run("WatchRuleset", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "/rulesets/a", r.URL.Path)
			fmt.Fprintf(w, `{"events": [{"type": "PUT", "path": "a"}], "revision": "rev"}`)
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ch := cli.WatchRulesets(ctx, "a")
		evs := <-ch
		require.NoError(t, evs.Err)
	})

	t.Run("WatchRuleset/NotFound", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ch := cli.WatchRulesets(ctx, "a")
		evs := <-ch
		require.Error(t, evs.Err)
	})

	t.Run("WatchRuleset/Errors", func(t *testing.T) {
		statuses := []int{
			http.StatusRequestTimeout,
			http.StatusInternalServerError,
			http.StatusBadRequest,
		}

		for _, status := range statuses {
			var i int32
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&i, 1) > 3 {
					w.WriteHeader(http.StatusOK)
					return
				}

				w.WriteHeader(status)
			}))
			defer ts.Close()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cli, err := client.New(ts.URL)
			require.NoError(t, err)
			cli.Logger = zerolog.New(ioutil.Discard)
			cli.WatchDelay = 1 * time.Millisecond

			ch := cli.WatchRulesets(ctx, "a")
			evs := <-ch
			require.NoError(t, evs.Err)
		}
	})

	t.Run("WatchRuleset/Cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-ctx.Done():
				return
			}
		}))
		defer ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ch := cli.WatchRulesets(ctx, "a")
		cancel()
		evs := <-ch
		require.Zero(t, evs)
	})
}
