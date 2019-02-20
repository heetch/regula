package regula_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
)

func ExampleRuleset() {
	rs, err := regula.NewRuleset(
		regula.NewSignature("string", nil),
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
		Signature: regula.NewSignature("string", nil),
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("some-string")),
		},
	})

	buf.Add("/path/to/string/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Signature: regula.NewSignature("string", nil),
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("some-string")),
		},
	})

	buf.Add("/path/to/int64/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Signature: regula.NewSignature("int64", nil),
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.Int64Value(10)),
		},
	})

	buf.Add("/path/to/float64/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Signature: regula.NewSignature("float64", nil),
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.Float64Value(3.14)),
		},
	})

	buf.Add("/path/to/bool/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Signature: regula.NewSignature("bool", nil),
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.BoolValue(true)),
		},
	})

	buf.Add("/path/to/duration/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Signature: regula.NewSignature("string", nil),
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("3s")),
		},
	})
}

func ExampleEngine() {
	engine := regula.NewEngine(ev)

	res, err := engine.Get(context.Background(), "/a/b/c", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		switch err {
		case errors.ErrRulesetNotFound:
			// when the ruleset doesn't exist
		case errors.ErrTypeMismatch:
			// when the ruleset returns the bad type
		case errors.ErrNoMatch:
			// when the ruleset doesn't match
		default:
			// something unexpected happened
		}
	}

	str, err := res.ToString()
	if err != nil {
		panic(err)
	}

	fmt.Println(str)
	fmt.Println(res.Version)
	// Output
	// some-string
	// 5b4cbdf307bb5346a6c42ac3
}

func ExampleEngine_GetBool() {
	engine := regula.NewEngine(ev)

	b, err := engine.GetBool(context.Background(), "/path/to/bool/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(b)
	// Output
	// true
}

func ExampleEngine_GetString() {
	engine := regula.NewEngine(ev)

	s, err := engine.GetString(context.Background(), "/path/to/string/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(s)
	// Output
	// some-string
}

func ExampleEngine_GetInt64() {
	engine := regula.NewEngine(ev)

	i, err := engine.GetInt64(context.Background(), "/path/to/int64/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(i)
	// Output
	// 10
}

func ExampleEngine_GetFloat64() {
	engine := regula.NewEngine(ev)

	f, err := engine.GetFloat64(context.Background(), "/path/to/float64/key", regula.Params{
		"product-id": "1234",
		"user-id":    "5678",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(f)
	// Output
	// 3.14
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
