package rules_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	ctx := context.Background()

	m := rules.MemoryGetter{Rulesets: map[string]*rule.Ruleset{
		"match-string-a": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("bar")), rule.ReturnsString("matched a")),
			},
		},
		"match-string-b": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsString("matched b")),
			},
		},
		"type-mismatch": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "int", Data: "5"}),
			},
		},
		"no-match": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.Eq(rule.StringValue("foo"), rule.StringValue("bar")), rule.ReturnsString("matched d")),
			},
		},
		"match-bool": &rule.Ruleset{
			Type: "bool",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "bool", Data: "true"}),
			},
		},
		"match-int64": &rule.Ruleset{
			Type: "int64",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "int64", Data: "-10"}),
			},
		},
		"match-float64": &rule.Ruleset{
			Type: "float64",
			Rules: []*rule.Rule{
				rule.New(rule.True(), &rule.Value{Type: "float64", Data: "-3.14"}),
			},
		},
		"match-duration": &rule.Ruleset{
			Type: "string",
			Rules: []*rule.Rule{
				rule.New(rule.True(), rule.ReturnsString("3s")),
			},
		},
	}}

	e := rules.NewEngine(&m)

	t.Run("LowLevel", func(t *testing.T) {
		str, err := e.GetString(ctx, "match-string-a", rule.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, "matched a", str)

		str, err = e.GetString(ctx, "match-string-b", nil)
		require.NoError(t, err)
		require.Equal(t, "matched b", str)

		b, err := e.GetBool(ctx, "match-bool", nil)
		require.NoError(t, err)
		require.True(t, b)

		i, err := e.GetInt64(ctx, "match-int64", nil)
		require.NoError(t, err)
		require.Equal(t, int64(-10), i)

		f, err := e.GetFloat64(ctx, "match-float64", nil)
		require.NoError(t, err)
		require.Equal(t, -3.14, f)

		_, err = e.GetString(ctx, "match-bool", nil)
		require.Equal(t, rules.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "type-mismatch", nil)
		require.Equal(t, rules.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "no-match", nil)
		require.Equal(t, rule.ErrNoMatch, err)

		_, err = e.GetString(ctx, "not-found", nil)
		require.Equal(t, rules.ErrRulesetNotFound, err)
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

var gt rules.Getter

func init() {
	gt = &rules.MemoryGetter{Rulesets: map[string]*rule.Ruleset{
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
	}}
}

func ExampleEngine() {
	engine := rules.NewEngine(gt)

	_, err := engine.GetString(context.Background(), "/a/b/c", rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		switch err {
		case rules.ErrRulesetNotFound:
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
	engine := rules.NewEngine(gt)

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
	engine := rules.NewEngine(gt)

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
	engine := rules.NewEngine(gt)

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
	engine := rules.NewEngine(gt)

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

	engine := rules.NewEngine(gt)

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
