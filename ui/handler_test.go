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

	"github.com/heetch/regula"
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

func TestPostNewRuleset(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		s := new(mock.RulesetService)
		rec := doRequest(NewHandler(s, ""), "POST", "/i/rulesets/",
			strings.NewReader(`{
			"path": "some-path",
			"signature": {
				"returnType": "string",
				"params": {
					"foo": "string",
					"bar": "int64"
				}
			}
}`))
		require.Equal(t, http.StatusCreated, rec.Code)
		require.Equal(t, 1, s.CreateCount)
	})

	t.Run("Errors", func(t *testing.T) {
		tests := []struct {
			name string
			err  error
			code int
		}{
			{"already exists", store.ErrAlreadyExists, http.StatusConflict},
			{"validation error", new(store.ValidationError), http.StatusBadRequest},
			{"unexpected error", errors.New("some error"), http.StatusInternalServerError},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				s := new(mock.RulesetService)
				s.CreateFn = func(_ context.Context, path string, _ *regula.Signature) error {
					return test.err
				}

				rec := doRequest(NewHandler(s, ""), "POST", "/i/rulesets/",
					strings.NewReader(`{
						"path": "some-path",
						"signature": {
							"returnType": "string",
							"params": {
								"foo": "string",
								"bar": "int64"
							}
						}
					}`))
				require.Equal(t, test.code, rec.Code)
				require.Equal(t, 1, s.CreateCount)
			})
		}

	})
}

func TestPUTNewRulesetVersionWithParserError(t *testing.T) {
	s := new(mock.RulesetService)
	s.GetFn = func(_ context.Context, path, _ string) (*regula.Ruleset, error) {
		return &regula.Ruleset{
			Path: path,
			Signature: &regula.Signature{
				ReturnType: "string",
				Params:     map[string]string{"foo": "string"},
			},
		}, nil
	}
	rec := doRequest(NewHandler(s, ""), "PUT", "/i/rulesets/some/path",
		strings.NewReader(`{
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

func TestPUTNewRulesetVersion(t *testing.T) {
	s := new(mock.RulesetService)
	s.GetFn = func(_ context.Context, path, _ string) (*regula.Ruleset, error) {
		return &regula.Ruleset{
			Path: path,
			Signature: &regula.Signature{
				ReturnType: "string",
				Params:     map[string]string{"foo": "string"},
			},
		}, nil
	}

	rec := doRequest(NewHandler(s, ""), "PUT", "/i/rulesets/some/path",
		strings.NewReader(`{
    "rules": [
        {
            "sExpr": "(= 1 1)",
            "returnValue": "wibble"
        }
    ]
}`))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, s.PutCount)
}

func TestInternalHandler(t *testing.T) {
	// this test checks if the handler deals with pagination correctly
	// and returns the right payload
	t.Run("OK", func(t *testing.T) {
		s := new(mock.RulesetService)

		// simulate a two page result
		s.ListFn = func(ctx context.Context, _ string, opt *store.ListOptions) (*store.Rulesets, error) {
			var rulesets store.Rulesets

			switch opt.ContinueToken {
			case "":
				for i := 0; i < 2; i++ {
					rulesets.Rulesets = append(rulesets.Rulesets, regula.Ruleset{
						Path: fmt.Sprintf("Path%d", i),
					})
				}

				rulesets.Continue = "continue"
			case "continue":
				for i := 2; i < 3; i++ {
					rulesets.Rulesets = append(rulesets.Rulesets, regula.Ruleset{
						Path: fmt.Sprintf("Path%d", i),
					})
				}
				rulesets.Continue = ""
			}

			return &rulesets, nil
		}

		rec := doRequest(NewHandler(s, ""), "GET", "/i/rulesets/", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `{"rulesets": [{"path": "Path0"},{"path": "Path1"},{"path": "Path2"}]}`, rec.Body.String())
	})

	t.Run("Empty result", func(t *testing.T) {
		s := new(mock.RulesetService)

		// simulate a two page result
		s.ListFn = func(ctx context.Context, _ string, opt *store.ListOptions) (*store.Rulesets, error) {
			return new(store.Rulesets), nil
		}

		rec := doRequest(NewHandler(s, ""), "GET", "/i/rulesets/", nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `{"rulesets": []}`, rec.Body.String())
	})

	// this test checks if the handler returns a 500 if a random error occurs in the ruleset service.
	t.Run("Error", func(t *testing.T) {
		s := new(mock.RulesetService)
		s.ListFn = func(ctx context.Context, _ string, opt *store.ListOptions) (*store.Rulesets, error) {
			return nil, errors.New("some error")
		}
	})
}

func TestConvertParams(t *testing.T) {
	cases := []struct {
		name    string
		input   map[string]string
		output  sexpr.Parameters
		errText string
	}{
		{
			name:  "single int64",
			input: map[string]string{"my-param": "int64"},
			output: sexpr.Parameters{
				"my-param": regrule.INTEGER,
			},
		},
		{
			name:  "single float64",
			input: map[string]string{"my-param": "float64"},
			output: sexpr.Parameters{
				"my-param": regrule.FLOAT,
			},
		},
		{
			name:  "single bool",
			input: map[string]string{"my-param": "bool"},
			output: sexpr.Parameters{
				"my-param": regrule.BOOLEAN,
			},
		},
		{
			name:  "single string",
			input: map[string]string{"my-param": "string"},
			output: sexpr.Parameters{
				"my-param": regrule.STRING,
			},
		},
		{
			name: "multiple parameters",
			input: map[string]string{
				"p1": "int64",
				"p2": "float64",
			},
			output: sexpr.Parameters{
				"p1": regrule.INTEGER,
				"p2": regrule.FLOAT,
			},
		},
		{
			name:    "no name error",
			input:   map[string]string{"": "int64"},
			errText: "parameter has no name",
		},
		{
			name:    "no type error",
			input:   map[string]string{"foo": ""},
			errText: "parameter (foo) has no type",
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
