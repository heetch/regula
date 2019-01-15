package regula_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/heetch/regula/errortype"
	"github.com/heetch/regula/rule"

	"github.com/heetch/regula"
)

func ExampleRuleset() {
	rs, err := regula.NewStringRuleset(
		rule.New(
			rule.Eq(
				rule.StringParam("group"),
				rule.StringValue("admin"),
			),
			rule.StringValue("first rule matched"),
		),
		rule.New(
			rule.In(
				rule.Int64Param("score"),
				rule.Int64Value(10),
				rule.Int64Value(20),
				rule.Int64Value(30),
			),
			rule.StringValue("second rule matched"),
		),
		rule.New(
			rule.True(),
			rule.StringValue("default rule matched"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	ret, err := rs.Eval(regula.Params{
		"group": "staff",
		"score": int64(20),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ret.Data)
	// Output
	// second rule matched
}

var ev regula.Evaluator

func init() {
	buf := regula.NewRulesetBuffer()
	ev = buf

	buf.Add("/a/b/c", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "string",
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("some-string")),
		},
	})

	buf.Add("/path/to/string/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "string",
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("some-string")),
		},
	})

	buf.Add("/path/to/int64/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "int64",
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.Int64Value(10)),
		},
	})

	buf.Add("/path/to/float64/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "float64",
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.Float64Value(3.14)),
		},
	})

	buf.Add("/path/to/bool/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "bool",
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.BoolValue(true)),
		},
	})

	buf.Add("/path/to/duration/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "string",
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("3s")),
		},
	})
}

func ExampleEngine() {
	engine := regula.NewEngine(ev)

	str, res, err := engine.GetString(context.Background(), "/a/b/c", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		switch err {
		case errortype.ErrRulesetNotFound:
			// when the ruleset doesn't exist
		case errortype.ErrTypeMismatch:
			// when the ruleset returns the bad type
		case errortype.ErrNoMatch:
			// when the ruleset doesn't match
		default:
			// something unexpected happened
		}
	}

	fmt.Println(str)
	fmt.Println(res.Version)
	// Output
	// some-string
	// 5b4cbdf307bb5346a6c42ac3
}

func ExampleEngine_GetBool() {
	engine := regula.NewEngine(ev)

	b, res, err := engine.GetBool(context.Background(), "/path/to/bool/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(b)
	fmt.Println(res.Version)
	// Output
	// true
	// 5b4cbdf307bb5346a6c42ac3
}

func ExampleEngine_GetString() {
	engine := regula.NewEngine(ev)

	s, res, err := engine.GetString(context.Background(), "/path/to/string/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)
	fmt.Println(res.Version)
	// Output
	// some-string
	// 5b4cbdf307bb5346a6c42ac3
}

func ExampleEngine_GetInt64() {
	engine := regula.NewEngine(ev)

	i, res, err := engine.GetInt64(context.Background(), "/path/to/int64/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(i)
	fmt.Println(res.Version)
	// Output
	// 10
	// 5b4cbdf307bb5346a6c42ac3
}

func ExampleEngine_GetFloat64() {
	engine := regula.NewEngine(ev)

	f, res, err := engine.GetFloat64(context.Background(), "/path/to/float64/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(f)
	fmt.Println(res.Version)
	// Output
	// 3.14
	// 5b4cbdf307bb5346a6c42ac3
}

func ExampleEngine_LoadStruct() {
	type Values struct {
		A string        `ruleset:"/path/to/string/key"`
		B int64         `ruleset:"/path/to/int64/key,required"`
		C time.Duration `ruleset:"/path/to/duration/key"`
	}

	var v Values

	engine := regula.NewEngine(ev)

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
