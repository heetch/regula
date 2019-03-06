package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heetch/regula/mock"
	regrule "github.com/heetch/regula/rule"
	"github.com/heetch/regula/rule/sexpr"
	"github.com/heetch/regula/store"
	"github.com/stretchr/testify/require"
)

func doRequest(h http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, body)
	h.ServeHTTP(w, r)
	return w
}

func TestPOSTNewRulesetWithParserError(t *testing.T) {
	s := new(mock.RulesetService)
	rec := doRequest(NewHandler(s, http.Dir("")), "POST", "/i/rulesets/",
		strings.NewReader(`{
    "path": "Path1",
    "signature": {
        "params": [
            {
                "name": "foo",
                "type": "string"
            }
        ],
        "returnType": "string"
    },
    "rules": [
        {
            "sExpr": "(= 1 1",
            "returnValue": "wibble"
        }
    ]
}`))
	require.Equal(t, http.StatusBadRequest, rec.Code)
	body := rec.Body.String()
	require.JSONEq(t, `{
    "error": "Error in rule 1: unexpected end of file",
    "fields": [
	{
	    "path": ["rules", "1", "sExpr"],
	    "error": {
		"message": "Error in rule 1: unexpected end of file",
		"line": 1,
		"char": 6,
		"absChar": 6
	 }
	}
    ]
}
`, body)
	require.Equal(t, 0, s.PutCount)
}

func TestPOSTNewRuleset(t *testing.T) {
	s := new(mock.RulesetService)
	rec := doRequest(NewHandler(s, http.Dir("")), "POST", "/i/rulesets/",
		strings.NewReader(`{
    "path": "Path1",
    "signature": {
        "params": [
            {
                "name": "foo",
                "type": "string"
            }
        ],
        "returnType": "string"
    },
    "rules": [
        {
            "sExpr": "(= 1 1)",
            "returnValue": "wibble"
        }
    ]
}`))
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, 1, s.PutCount)
}

func TestInternalHandler(t *testing.T) {
	// this test checks if the handler deals with pagination correctly
	// and returns the right payload
	t.Run("OK", func(t *testing.T) {
		s := new(mock.RulesetService)

		// simulate a two page result
		s.ListFn = func(ctx context.Context, _ string, opt *store.ListOptions) (*store.RulesetEntries, error) {
			var entries store.RulesetEntries

			switch opt.ContinueToken {
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

		rec := doRequest(NewHandler(s, http.Dir("")), "GET", "/i/rulesets/", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `{"rulesets": [{"path": "Path0"},{"path": "Path1"},{"path": "Path2"}]}`, rec.Body.String())
	})

	t.Run("Empty result", func(t *testing.T) {
		s := new(mock.RulesetService)

		// simulate a two page result
		s.ListFn = func(ctx context.Context, _ string, opt *store.ListOptions) (*store.RulesetEntries, error) {
			return new(store.RulesetEntries), nil
		}

		rec := doRequest(NewHandler(s, http.Dir("")), "GET", "/i/rulesets/", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `{"rulesets": []}`, rec.Body.String())
	})

	// this test checks if the handler returns a 500 if a random error occurs in the ruleset service.
	t.Run("Error", func(t *testing.T) {
		s := new(mock.RulesetService)
		s.ListFn = func(ctx context.Context, _ string, opt *store.ListOptions) (*store.RulesetEntries, error) {
			return nil, errors.New("some error")
		}
	})
}

func TestConvertParams(t *testing.T) {
	cases := []struct {
		name    string
		input   []param
		output  sexpr.Parameters
		errText string
	}{
		{
			name:  "single int64",
			input: []param{{"name": "my-param", "type": "int64"}},
			output: sexpr.Parameters{
				"my-param": regrule.INTEGER,
			},
		},
		{
			name:  "single float64",
			input: []param{{"name": "my-param", "type": "float64"}},
			output: sexpr.Parameters{
				"my-param": regrule.FLOAT,
			},
		},
		{
			name:  "single bool",
			input: []param{{"name": "my-param", "type": "bool"}},
			output: sexpr.Parameters{
				"my-param": regrule.BOOLEAN,
			},
		},
		{
			name:  "single string",
			input: []param{{"name": "my-param", "type": "string"}},
			output: sexpr.Parameters{
				"my-param": regrule.STRING,
			},
		},
		{
			name: "multiple parameters",
			input: []param{
				{"name": "p1", "type": "int64"},
				{"name": "p2", "type": "float64"},
			},
			output: sexpr.Parameters{
				"p1": regrule.INTEGER,
				"p2": regrule.FLOAT,
			},
		},
		{
			name:    "no name error",
			input:   []param{{"type": "int64"}},
			errText: "parameter 0 has no name",
		},
		{
			name:    "no type error",
			input:   []param{{"name": "foo"}},
			errText: "parameter 0 (foo) has no type",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := convertParams(c.input)
			if c.errText != "" {
				require.EqualError(t, err, c.errText)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.output, result)

		})
	}
}
