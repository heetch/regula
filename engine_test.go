package rules_test

import (
	"fmt"
	"log"
	"path"
	"testing"
	"time"

	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
	"github.com/heetch/rules-engine/store"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	namespace string
	ruleSets  map[string]*rule.Ruleset
}

func newMockStore(namespace string, ruleSets map[string]*rule.Ruleset) *mockStore {
	return &mockStore{
		namespace: namespace,
		ruleSets:  ruleSets,
	}
}

func (s *mockStore) Get(key string) (*rule.Ruleset, error) {
	key = path.Join("/", key)

	rs, ok := s.ruleSets[key]
	if !ok {
		err := store.ErrRulesetNotFound
		return nil, err
	}

	return rs, nil
}

func TestEngine(t *testing.T) {
	m := newMockStore("/rules", map[string]*rule.Ruleset{
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
		str, err := e.GetString("/match-string-a", rule.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, "matched a", str)

		str, err = e.GetString("/match-string-b", nil)
		require.NoError(t, err)
		require.Equal(t, "matched b", str)

		b, err := e.GetBool("/match-bool", nil)
		require.NoError(t, err)
		require.True(t, b)

		i, err := e.GetInt64("/match-int64", nil)
		require.NoError(t, err)
		require.Equal(t, int64(-10), i)

		f, err := e.GetFloat64("/match-float64", nil)
		require.NoError(t, err)
		require.Equal(t, -3.14, f)

		_, err = e.GetString("/match-bool", nil)
		require.Equal(t, rules.ErrTypeMismatch, err)

		_, err = e.GetString("/type-mismatch", nil)
		require.Equal(t, rules.ErrTypeMismatch, err)

		_, err = e.GetString("/no-match", nil)
		require.Equal(t, rule.ErrNoMatch, err)

		_, err = e.GetString("/not-found", nil)
		require.Equal(t, store.ErrRulesetNotFound, err)
	})

	t.Run("StructLoading", func(t *testing.T) {
		to := struct {
			StringA  string        `ruleset:"match-string-a"`
			Bool     bool          `ruleset:"match-bool"`
			Int64    int64         `ruleset:"match-int64"`
			Float64  float64       `ruleset:"match-float64"`
			Duration time.Duration `ruleset:"match-duration"`
		}{}

		err := e.LoadStruct(&to, rule.Params{
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
			StringA string `ruleset:"match-string-a"`
			Wrong   string `ruleset:"no-exists"`
		}{}

		err := e.LoadStruct(&to, rule.Params{
			"foo": "bar",
		})

		require.Equal(t, store.ErrRulesetNotFound, err)
	})

	t.Run("StructLoadingMissingParam", func(t *testing.T) {
		to := struct {
			StringA string `ruleset:"match-string-a"`
		}{}

		err := e.LoadStruct(&to, nil)

		require.Error(t, err)
	})
}

var st store.Store

func init() {
	st = newMockStore("/", map[string]*rule.Ruleset{
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
	engine := rules.NewEngine(st)

	_, err := engine.GetString("/a/b/c", rule.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		switch err {
		case store.ErrRulesetNotFound:
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
	engine := rules.NewEngine(st)

	b, err := engine.GetBool("/path/to/bool/key", rule.Params{
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
	engine := rules.NewEngine(st)

	s, err := engine.GetString("/path/to/string/key", rule.Params{
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
	engine := rules.NewEngine(st)

	s, err := engine.GetInt64("/path/to/int64/key", rule.Params{
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
	engine := rules.NewEngine(st)

	f, err := engine.GetFloat64("/path/to/float64/key", rule.Params{
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

	engine := rules.NewEngine(st)

	err := engine.LoadStruct(&v, rule.Params{
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
