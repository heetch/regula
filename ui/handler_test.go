package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heetch/regula/mock"
	"github.com/heetch/regula/store"
	"github.com/stretchr/testify/require"
)

func doRequest(h http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, body)
	h.ServeHTTP(w, r)
	return w
}

func TestInternalHandler(t *testing.T) {
	t.Run("Rulesets", func(t *testing.T) {
		// this test checks if the handler deals with pagination correctly
		// and returns the right payload
		t.Run("OK", func(t *testing.T) {
			s := new(mock.RulesetService)

			// simulate a two page result
			s.ListPathsFn = func(ctx context.Context, _ string, limit int, token string) (*store.RulesetEntries, error) {
				var entries store.RulesetEntries

				switch token {
				case "":
					for i := 0; i < 2; i++ {
						entries.Entries = append(entries.Entries, store.RulesetEntry{
							Path: fmt.Sprintf("Path%d", i),
						})
					}

					entries.Continue = "continue"
				case "continue":
					for i := 2; i < 3; i++ {
						entries.Entries = append(entries.Entries, store.RulesetEntry{
							Path: fmt.Sprintf("Path%d", i),
						})
					}
					entries.Continue = ""
				}

				return &entries, nil
			}

			rec := doRequest(NewHandler(s, ""), "GET", "/i/rulesets/", nil)
			require.Equal(t, http.StatusOK, rec.Code)
			require.JSONEq(t, `{"rulesets": [{"path": "Path0"},{"path": "Path1"},{"path": "Path2"}]}`, rec.Body.String())
		})

		// this test checks if the handler returns a 500 if a random error occurs in the ruleset service.
		t.Run("Error", func(t *testing.T) {
			s := new(mock.RulesetService)
			s.ListPathsFn = func(ctx context.Context, _ string, limit int, token string) (*store.RulesetEntries, error) {
				return nil, errors.New("some error")
			}

			rec := doRequest(NewHandler(s, ""), "GET", "/i/rulesets/", nil)
			require.Equal(t, http.StatusInternalServerError, rec.Code)
		})
	})
}
