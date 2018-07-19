package regula_test

import (
	"fmt"
	"log"

	"github.com/heetch/regula"
)

func Example() {
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
