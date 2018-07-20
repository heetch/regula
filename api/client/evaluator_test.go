package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestEvaluator(t *testing.T) {
	t.Run("Watch disabled", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := r.URL.Query()["list"]; ok {
				fmt.Fprintf(w, `{"revision": "revA", "rulesets": [{"path": "a", "version":"1"}]}`)
				return
			}

			t.Error("shouldn't reach this part")
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cli, err := New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ev, err := NewEvaluator(ctx, cli, "a", false)
		require.NoError(t, err)

		err = ev.Close()
		require.NoError(t, err)

		_, version, err := ev.Latest("a")
		require.NoError(t, err)
		require.Equal(t, "1", version)
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

		cli, err := New(ts.URL)
		require.NoError(t, err)
		cli.Logger = zerolog.New(ioutil.Discard)

		ev, err := NewEvaluator(ctx, cli, "a", true)
		require.NoError(t, err)

		<-didWatch
		err = ev.Close()
		require.NoError(t, err)

		_, version, err := ev.Latest("a")
		require.NoError(t, err)
		require.Equal(t, "2", version)
	})
}
