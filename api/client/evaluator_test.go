package client_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api/client"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluator(t *testing.T) {
	t.Run("Watch disabled", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := r.URL.Query()["list"]; ok {
				if continueToken := r.URL.Query().Get("continue"); continueToken != "" {
					assert.Equal(t, "some-token", continueToken)
					fmt.Fprintf(w, `{"revision": "revB", "rulesets": [{"path": "a", "version":"2"}]}`)
					return
				}

				fmt.Fprintf(w, `{"revision": "revA", "rulesets": [{"path": "a", "version":"1"}], "continue": "some-token"}`)
				return
			}

			t.Error("shouldn't reach this part")
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ev, err := client.NewEvaluator(ctx, cli, "a", false)
		require.NoError(t, err)

		err = ev.Close()
		require.NoError(t, err)

		_, version, err := ev.Latest("a")
		require.NoError(t, err)
		require.Equal(t, "2", version)
	})

	t.Run("Watch enabled", func(t *testing.T) {
		watchCount := 0
		didWatch := make(chan struct{})

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := r.URL.Query()["list"]; ok {
				fmt.Fprintf(w, `{"revision": "revA", "rulesets": [{"path": "a", "version":"1"}]}`)
				return
			}

			watchCount++

			if watchCount > 1 {
				close(didWatch)
				return
			}

			fmt.Fprintf(w, `{"events": [{"type": "PUT", "path": "a", "version": "2"}], "revision": "revB"}`)
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := client.New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ev, err := client.NewEvaluator(ctx, cli, "a", true)
		require.NoError(t, err)

		<-didWatch
		err = ev.Close()
		require.NoError(t, err)

		_, version, err := ev.Latest("a")
		require.NoError(t, err)
		require.Equal(t, "2", version)
	})
}

var (
	addr string
)

func ExampleEvaluator_withoutWatch() {
	cli, err := client.New(addr)
	if err != nil {
		log.Fatal(err)
	}

	ev, err := client.NewEvaluator(context.Background(), cli, "prefix", false)
	if err != nil {
		log.Fatal(err)
	}

	ng := regula.NewEngine(ev)
	str, err := ng.GetString(context.Background(), "some/path", regula.Params{
		"id": "123",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)
}

func ExampleEvaluator_withWatch() {
	cli, err := client.New(addr)
	if err != nil {
		log.Fatal(err)
	}

	ev, err := client.NewEvaluator(context.Background(), cli, "prefix", true)
	if err != nil {
		log.Fatal(err)
	}
	// stopping the watcher
	defer ev.Close()

	ng := regula.NewEngine(ev)
	str, err := ng.GetString(context.Background(), "some/path", regula.Params{
		"id": "123",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)
}
