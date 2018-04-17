package rules_test

import (
	"context"
	"fmt"
	"log"
	"path"
	"testing"
	"time"

	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/client"
	"github.com/heetch/rules-engine/rule"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	namespace string
	ruleSets  map[string]*rule.Ruleset
}

func newMockClient(namespace string, ruleSets map[string]*rule.Ruleset) *mockClient {
	return &mockClient{
		namespace: namespace,
		ruleSets:  ruleSets,
	}
}

func (s *mockClient) Get(ctx context.Context, key string) (*rule.Ruleset, error) {
	key = path.Join("/", key)

	rs, ok := s.ruleSets[key]
	if !ok {
		err := client.ErrRulesetNotFound
		return nil, err
	}

	return rs, nil
}

func TestEngine(t *testing.T) {
	ctx := context.Background()

	m := newMockClient("/rules", map[string]*rule.Ruleset{
		"/match-string-a": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("bar")), rule.ReturnsString("matched a")),
			},
		},
		"/match-string-b": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsString("matched b")),
			},
		},
		"/type-mismatch": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "int", Data: "5"}),
			},
		},
		"/no-match": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.Eq(rule.StringValue("foo"), rule.StringValue("bar")), rule.ReturnsString("matched d")),
			},
		},
		"/match-bool": &rule.Ruleset{
			Type: "bool",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "bool", Data: "true"}),
			},
		},
		"/match-int64": &rule.Ruleset{
			Type: "int64",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "int64", Data: "-10"}),
			},
		},
		"/match-float64": &rule.Ruleset{
			Type: "float64",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "float64", Data: "-3.14"}),
			},
		},
		"/match-duration": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsString("3s")),
			},
		},
	})

	e := rules.NewEngine(m)

	t.Run("LowLevel", func(t *testing.T) {
		str, err := e.GetString(ctx, "/match-string-a", rule.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, "matched a", str)

		str, err = e.GetString(ctx, "/match-string-b", nil)
		require.NoError(t, err)
		require.Equal(t, "matched b", str)

		b, err := e.GetBool(ctx, "/match-bool", nil)
		require.NoError(t, err)
		require.True(t, b)

		i, err := e.GetInt64(ctx, "/match-int64", nil)
		require.NoError(t, err)
		require.Equal(t, int64(-10), i)

		f, err := e.GetFloat64(ctx, "/match-float64", nil)
		require.NoError(t, err)
		require.Equal(t, -3.14, f)

		_, err = e.GetString(ctx, "/match-bool", nil)
		require.Equal(t, rules.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "/type-mismatch", nil)
		require.Equal(t, rules.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "/no-match", nil)
		require.Equal(t, rule.ErrNoMatch, err)

		_, err = e.GetString(ctx, "/not-found", nil)
		require.Equal(t, client.ErrRulesetNotFound, err)
	})

	t.Run("StructLoading", func(t *testing.T) {
		to := struct {
			StringA  string        `ruleset:"match-string-a"`
			Bool     bool          `ruleset:"match-bool"`
			Int64    int64         `ruleset:"match-int64"`
			Float64  float64       `ruleset:"match-float64"`
			Duration time.Duration `ruleset:"match-duration"`
		}{}

		err := e.LoadStruct(ctx, &to, rule.Params{
			"foo": "bar",
		})

		require.NoError(t, err)
		require.Equal(t, "matched a", to.StringA)
		require.Equal(t, true, to.Bool)
		require.Equal(t, int64(-10), to.Int64)
		require.Equal(t, -3.14, to.Float64)
		require.Equal(t, 3*time.Second, to.Duration)
	})

	t.Run("StructLoadingWrongKey", func(t *testing.T) {
		to := struct {
			StringA string `ruleset:"match-string-a,required"`
			Wrong   string `ruleset:"no-exists,required"`
		}{}

		err := e.LoadStruct(ctx, &to, rule.Params{
			"foo": "bar",
		})

		require.Error(t, err)
	})

	t.Run("StructLoadingMissingParam", func(t *testing.T) {
		to := struct {
			StringA string `ruleset:"match-string-a"`
		}{}

		err := e.LoadStruct(ctx, &to, nil)

		require.Error(t, err)
	})
}

var cli client.Client

func init() {
	cli = newMockClient("/", map[string]*rule.Ruleset{
		"/path/to/string/key": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsString("some-string")),
			},
		},

		"/path/to/int64/key": &rule.Ruleset{
			Type: "int64",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsInt64(10)),
			},
		},

		"/path/to/float64/key": &rule.Ruleset{
			Type: "float64",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsFloat64(3.14)),
			},
		},

		"/path/to/bool/key": &rule.Ruleset{
			Type: "bool",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsBool(true)),
			},
		},

		"/path/to/duration/key": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsString("3s")),
			},
		},
	})
}

func ExampleEngine() {
	engine := rules.NewEngine(cli)

	_, err := engine.GetString(context.Background(), "/a/b/c", rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		switch err {
		case client.ErrRulesetNotFound:
			// when the ruleset doesn't exist
		case rules.ErrTypeMismatch:
			// when the ruleset returns the bad type
		case rule.ErrNoMatch:
			// when the ruleset doesn't match
		default:
			// something unexpected happened
		}
	}
}

func ExampleEngine_GetBool() {
	engine := rules.NewEngine(cli)

	b, err := engine.GetBool(context.Background(), "/path/to/bool/key", rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(b)
	// Output: true
}

func ExampleEngine_GetString() {
	engine := rules.NewEngine(cli)

	s, err := engine.GetString(context.Background(), "/path/to/string/key", rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)
	// Output: some-string
}

func ExampleEngine_GetInt64() {
	engine := rules.NewEngine(cli)

	s, err := engine.GetInt64(context.Background(), "/path/to/int64/key", rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)
	// Output: 10
}

func ExampleEngine_GetFloat64() {
	engine := rules.NewEngine(cli)

	f, err := engine.GetFloat64(context.Background(), "/path/to/float64/key", rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(f)
	// Output: 3.14
}

func ExampleEngine_LoadStruct() {
	type Values struct {
		A string        `ruleset:"/path/to/string/key"`
		B int64         `ruleset:"/path/to/int64/key,required"`
		C time.Duration `ruleset:"/path/to/duration/key"`
	}

	var v Values

	engine := rules.NewEngine(cli)

	err := engine.LoadStruct(context.Background(), &v, rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(v.A)
	fmt.Println(v.B)
	fmt.Println(v.C)
	// Output:
	// some-string
	// 10
	// 3s
}
