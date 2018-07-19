package regula_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/heetch/regula"
	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	ctx := context.Background()

	var m regula.MemoryGetter

	m.AddRuleset("match-string-a", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.Eq(regula.StringParam("foo"), regula.StringValue("bar")), regula.ReturnsString("matched a v1")),
		},
	})
	m.AddRuleset("match-string-a", "2", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.Eq(regula.StringParam("foo"), regula.StringValue("bar")), regula.ReturnsString("matched a v2")),
		},
	})
	m.AddRuleset("match-string-b", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("matched b")),
		},
	})
	m.AddRuleset("type-mismatch", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "int", Data: "5"}),
		},
	})
	m.AddRuleset("no-match", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.Eq(regula.StringValue("foo"), regula.StringValue("bar")), regula.ReturnsString("matched d")),
		},
	})
	m.AddRuleset("match-bool", "1", &regula.Ruleset{
		Type: "bool",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "bool", Data: "true"}),
		},
	})
	m.AddRuleset("match-int64", "1", &regula.Ruleset{
		Type: "int64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "int64", Data: "-10"}),
		},
	})
	m.AddRuleset("match-float64", "1", &regula.Ruleset{
		Type: "float64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), &regula.Value{Type: "float64", Data: "-3.14"}),
		},
	})
	m.AddRuleset("match-duration", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("3s")),
		},
	})

	e := regula.NewEngine(&m)

	t.Run("LowLevel", func(t *testing.T) {
		str, err := e.GetString(ctx, "match-string-a", regula.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, "matched a v2", str)

		str, err = e.GetString(ctx, "match-string-a", regula.Params{
			"foo": "bar",
		}, regula.Version("1"))
		require.NoError(t, err)
		require.Equal(t, "matched a v1", str)

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
		require.Equal(t, regula.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "type-mismatch", nil)
		require.Equal(t, regula.ErrTypeMismatch, err)

		_, err = e.GetString(ctx, "no-match", nil)
		require.Equal(t, regula.ErrNoMatch, err)

		_, err = e.GetString(ctx, "not-found", nil)
		require.Equal(t, regula.ErrRulesetNotFound, err)
	})

	t.Run("StructLoading", func(t *testing.T) {
		to := struct {
			StringA  string        `ruleset:"match-string-a"`
			Bool     bool          `ruleset:"match-bool"`
			Int64    int64         `ruleset:"match-int64"`
			Float64  float64       `ruleset:"match-float64"`
			Duration time.Duration `ruleset:"match-duration"`
		}{}

		err := e.LoadStruct(ctx, &to, regula.Params{
			"foo": "bar",
		})

		require.NoError(t, err)
		require.Equal(t, "matched a v2", to.StringA)
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

		err := e.LoadStruct(ctx, &to, regula.Params{
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

var gt regula.Getter

func init() {
	var m regula.MemoryGetter
	gt = &m

	m.AddRuleset("/path/to/string/key", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("some-string")),
		},
	})

	m.AddRuleset("/path/to/int64/key", "1", &regula.Ruleset{
		Type: "int64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsInt64(10)),
		},
	})

	m.AddRuleset("/path/to/float64/key", "1", &regula.Ruleset{
		Type: "float64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsFloat64(3.14)),
		},
	})

	m.AddRuleset("/path/to/bool/key", "1", &regula.Ruleset{
		Type: "bool",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsBool(true)),
		},
	})

	m.AddRuleset("/path/to/duration/key", "1", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("3s")),
		},
	})
}

func ExampleEngine() {
	engine := regula.NewEngine(gt)

	_, err := engine.GetString(context.Background(), "/a/b/c", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		switch err {
		case regula.ErrRulesetNotFound:
			// when the ruleset doesn't exist
		case regula.ErrTypeMismatch:
			// when the ruleset returns the bad type
		case regula.ErrNoMatch:
			// when the ruleset doesn't match
		default:
			// something unexpected happened
		}
	}
}

func ExampleEngine_GetBool() {
	engine := regula.NewEngine(gt)

	b, err := engine.GetBool(context.Background(), "/path/to/bool/key", regula.Params{
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
	engine := regula.NewEngine(gt)

	s, err := engine.GetString(context.Background(), "/path/to/string/key", regula.Params{
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
	engine := regula.NewEngine(gt)

	s, err := engine.GetInt64(context.Background(), "/path/to/int64/key", regula.Params{
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
	engine := regula.NewEngine(gt)

	f, err := engine.GetFloat64(context.Background(), "/path/to/float64/key", regula.Params{
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

	engine := regula.NewEngine(gt)

	err := engine.LoadStruct(context.Background(), &v, regula.Params{
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
