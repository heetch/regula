package rules_test

import (
	"fmt"
	"log"
	"testing"

	rules "github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
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
	rs, ok := s.ruleSets[key]
	if !ok {
		err := rules.ErrRulesetNotFound
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
	})

	e := rules.NewEngine(m)
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
	require.Equal(t, rules.ErrRulesetNotFound, err)
}

var store rules.Store

func init() {
	store = newMockStore("/", map[string]*rule.Ruleset{
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
	})
}

func ExampleEngine() {
	engine := rules.NewEngine(store)

	_, err := engine.GetString("/a/b/c", rule.Params{
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
	engine := rules.NewEngine(store)

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
	engine := rules.NewEngine(store)

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
	engine := rules.NewEngine(store)

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
	engine := rules.NewEngine(store)

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
