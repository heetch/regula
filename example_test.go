package regula_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/heetch/regula"
)

func ExampleRule() {
	r := regula.NewRule(
		regula.Eq(
			regula.StringValue("foo"),
			regula.StringParam("bar"),
		),
		regula.ReturnsString("matched"),
	)

	ret, err := r.Eval(regula.Params{
		"bar": "foo",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ret.Data)
	// Output
	// matched
}

func ExampleAnd() {
	tree := regula.And(
		regula.Eq(
			regula.Int64Value(10),
			regula.Int64Param("foo"),
		),
		regula.Not(
			regula.Eq(
				regula.Float64Value(1.5),
				regula.Float64Param("bar"),
			),
		),
	)

	val, err := tree.Eval(regula.Params{
		"foo": int64(10),
		"bar": 1.6,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleOr() {
	tree := regula.Or(
		regula.Eq(
			regula.Float64Value(1.2),
			regula.Float64Param("foo"),
		),
		regula.Eq(
			regula.Float64Value(3.14),
			regula.Float64Param("foo"),
		),
	)

	val, err := tree.Eval(regula.Params{
		"foo": 3.14,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_string() {
	tree := regula.Eq(
		regula.StringValue("bar"),
		regula.StringValue("bar"),
		regula.StringParam("foo"),
	)

	val, err := tree.Eval(regula.Params{
		"foo": "bar",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_bool() {
	tree := regula.Eq(
		regula.BoolValue(false),
		regula.Not(regula.BoolValue(true)),
		regula.Eq(
			regula.StringValue("bar"),
			regula.StringValue("baz"),
		),
	)

	val, err := tree.Eval(regula.Params{
		"foo": "bar",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_int64() {
	tree := regula.Eq(
		regula.Int64Value(10),
		regula.Int64Param("foo"),
	)

	val, err := tree.Eval(regula.Params{
		"foo": int64(10),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_float64() {
	tree := regula.Eq(
		regula.Float64Value(3.14),
		regula.Float64Param("foo"),
	)

	val, err := tree.Eval(regula.Params{
		"foo": 3.14,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleIn() {
	tree := regula.In(
		regula.StringValue("c"),
		regula.StringValue("a"),
		regula.StringValue("b"),
		regula.StringValue("c"),
		regula.StringValue("d"),
	)

	val, err := tree.Eval(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleNot() {
	tree := regula.Not(regula.BoolValue(false))

	val, err := tree.Eval(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleStringParam() {
	tree := regula.StringParam("foo")

	val, err := tree.Eval(regula.Params{
		"foo": "bar",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: bar
}

func ExampleTrue() {
	tree := regula.True()

	val, err := tree.Eval(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleRuleset() {
	rs, err := regula.NewStringRuleset(
		regula.NewRule(
			regula.Eq(
				regula.StringParam("group"),
				regula.StringValue("admin"),
			),
			regula.ReturnsString("first rule matched"),
		),
		regula.NewRule(
			regula.In(
				regula.Int64Param("score"),
				regula.Int64Value(10),
				regula.Int64Value(20),
				regula.Int64Value(30),
			),
			regula.ReturnsString("second rule matched"),
		),
		regula.NewRule(
			regula.True(),
			regula.ReturnsString("default rule matched"),
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
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("some-string")),
		},
	})

	buf.Add("/path/to/string/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("some-string")),
		},
	})

	buf.Add("/path/to/int64/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "int64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsInt64(10)),
		},
	})

	buf.Add("/path/to/float64/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "float64",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsFloat64(3.14)),
		},
	})

	buf.Add("/path/to/bool/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "bool",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsBool(true)),
		},
	})

	buf.Add("/path/to/duration/key", "5b4cbdf307bb5346a6c42ac3", &regula.Ruleset{
		Type: "string",
		Rules: []*regula.Rule{
			regula.NewRule(regula.True(), regula.ReturnsString("3s")),
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
