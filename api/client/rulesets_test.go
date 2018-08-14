package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/api/client"
	"github.com/heetch/regula/rule"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ev regula.Evaluator = new(client.RulesetService)

func ExampleRulesetService_List() {
	c, err := client.New("http://127.0.0.1:5331")
	if err != nil {
		log.Fatal(err)
	}

	list, err := c.Rulesets.List(context.Background(), "prefix", nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range list.Rulesets {
		e.Ruleset.Eval(nil)
	}
}

func ExampleRulesetService_List_withPagination() {
	c, err := client.New("http://127.0.0.1:5331")
	if err != nil {
		log.Fatal(err)
	}

	list, err := c.Rulesets.List(context.Background(), "prefix", &client.ListOptions{
		Limit: 20,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range list.Rulesets {
		e.Ruleset.Eval(nil)
	}

	for list.Continue != "" {
		list, err = c.Rulesets.List(context.Background(), "prefix", &client.ListOptions{
			Limit:    20,
			Continue: list.Continue,
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, e := range list.Rulesets {
			e.Ruleset.Eval(nil)
		}
	}
}

func ExampleRulesetService_Eval() {
	c, err := client.New("http://127.0.0.1:5331")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Rulesets.Eval(context.Background(), "path/to/ruleset", regula.Params{
		"foo": "bar",
		"baz": int64(42),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Value.Data)
	fmt.Println(resp.Value.Type)
	fmt.Println(resp.Version)
}

func ExampleRulesetService_EvalVersion() {
	c, err := client.New("http://127.0.0.1:5331")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Rulesets.EvalVersion(context.Background(), "path/to/ruleset", "xyzabc", regula.Params{
		"foo": "bar",
		"baz": int64(42),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Value.Data)
	fmt.Println(resp.Value.Type)
	fmt.Println(resp.Version)
}

func TestRulesetService(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "some err"}`)
		}))
		defer ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		_, err = cli.Rulesets.List(context.Background(), "", nil)
		aerr := err.(*api.Error)
		require.Equal(t, "some err", aerr.Err)
	})

	t.Run("Retries/Success", func(t *testing.T) {
		i := 0

		rs, err := regula.NewInt64Ruleset(rule.New(rule.True(), rule.Int64Value(1)))
		require.NoError(t, err)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			i++

			var body regula.Ruleset
			err := json.NewDecoder(r.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, rs, &body)

			if i < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "some err"}`)
		}))
		defer ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)
		cli.RetryDelay = 10 * time.Millisecond

		_, err = cli.Rulesets.Put(context.Background(), "path", rs)
		aerr := err.(*api.Error)
		require.Equal(t, "some err", aerr.Err)
	})

	t.Run("Retries/Failure", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)
		cli.RetryDelay = 10 * time.Millisecond

		rs, err := regula.NewInt64Ruleset(rule.New(rule.True(), rule.Int64Value(1)))
		require.NoError(t, err)

		_, err = cli.Rulesets.Put(context.Background(), "path", rs)
		aerr := err.(*api.Error)
		require.Equal(t, http.StatusInternalServerError, aerr.Response.StatusCode)
	})

	t.Run("Retries/NetworkError", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		// close immediately
		ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)
		cli.RetryDelay = 10 * time.Millisecond

		rs, err := regula.NewInt64Ruleset(rule.New(rule.True(), rule.Int64Value(1)))
		require.NoError(t, err)

		_, err = cli.Rulesets.Put(context.Background(), "path", rs)
		_, ok := err.(net.Error)
		require.True(t, ok)
	})

	t.Run("ListRulesets", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Contains(t, r.URL.Query(), "list")
			assert.Equal(t, "some-token", r.URL.Query().Get("continue"))
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "/rulesets/prefix", r.URL.Path)
			fmt.Fprintf(w, `{"revision": "rev", "rulesets": [{"path": "a"}]}`)
		}))
		defer ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		rs, err := cli.Rulesets.List(context.Background(), "prefix", &client.ListOptions{
			Limit:    10,
			Continue: "some-token",
		})
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
			fmt.Fprintf(w, `{"value": {"data": "baz", "type": "string", "kind": "value"}, "version": "1234"}`)
		}))
		defer ts.Close()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		exp := regula.EvalResult{Value: rule.StringValue("baz"), Version: "1234"}

		resp, err := cli.Rulesets.Eval(context.Background(), "path/to/ruleset", regula.Params{
			"foo": "bar",
		})
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

		rs, err := regula.NewInt64Ruleset(rule.New(rule.True(), rule.Int64Value(1)))
		require.NoError(t, err)

		ars, err := cli.Rulesets.Put(context.Background(), "a", rs)
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

		ch := cli.Rulesets.Watch(ctx, "a", "")
		evs := <-ch
		require.NoError(t, evs.Err)
	})

	t.Run("WatchRuleset/Revision", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "rev", r.URL.Query().Get("revision"))
			assert.Equal(t, "/rulesets/a", r.URL.Path)
			fmt.Fprintf(w, `{"events": [{"type": "PUT", "path": "a"}], "revision": "rev"}`)
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ch := cli.Rulesets.Watch(ctx, "a", "rev")
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

		ch := cli.Rulesets.Watch(ctx, "a", "")
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

			ch := cli.Rulesets.Watch(ctx, "a", "")
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

		ch := cli.Rulesets.Watch(ctx, "a", "")
		cancel()
		evs := <-ch
		require.Zero(t, evs)
	})
}
