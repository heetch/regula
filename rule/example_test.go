package rule_test

import (
	"fmt"
	"log"

	"github.com/heetch/rules-engine/rule"
)

func Example() {
	r := rule.New(
		rule.Eq(
			rule.StringValue("foo"),
			rule.StringParam("bar"),
		),
		rule.ReturnsString("matched"),
	)

	ret, err := r.Eval(rule.Params{
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
	tree := rule.And(
		rule.Eq(
			rule.Int64Value(10),
			rule.Int64Param("foo"),
		),
		rule.Not(
			rule.Eq(
				rule.Float64Value(1.5),
				rule.Float64Param("bar"),
			),
		),
	)

	val, err := tree.Eval(rule.Params{
		"foo": 10,
		"bar": 1.6,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleOr() {
	tree := rule.Or(
		rule.Eq(
			rule.Float64Value(1.2),
			rule.Float64Param("foo"),
		),
		rule.Eq(
			rule.Float64Value(3.14),
			rule.Float64Param("foo"),
		),
	)

	val, err := tree.Eval(rule.Params{
		"foo": 3.14,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_string() {
	tree := rule.Eq(
		rule.StringValue("bar"),
		rule.StringValue("bar"),
		rule.StringParam("foo"),
	)

	val, err := tree.Eval(rule.Params{
		"foo": "bar",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_bool() {
	tree := rule.Eq(
		rule.BoolValue(false),
		rule.Not(rule.BoolValue(true)),
		rule.Eq(
			rule.StringValue("bar"),
			rule.StringValue("baz"),
		),
	)

	val, err := tree.Eval(rule.Params{
		"foo": "bar",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_int64() {
	tree := rule.Eq(
		rule.Int64Value(10),
		rule.Int64Param("foo"),
	)

	val, err := tree.Eval(rule.Params{
		"foo": 10,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleEq_float64() {
	tree := rule.Eq(
		rule.Float64Value(3.14),
		rule.Float64Param("foo"),
	)

	val, err := tree.Eval(rule.Params{
		"foo": 3.14,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleIn() {
	tree := rule.In(
		rule.StringValue("c"),
		rule.StringValue("a"),
		rule.StringValue("b"),
		rule.StringValue("c"),
		rule.StringValue("d"),
	)

	val, err := tree.Eval(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleNot() {
	tree := rule.Not(rule.BoolValue(false))

	val, err := tree.Eval(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}

func ExampleStringParam() {
	tree := rule.StringParam("foo")

	val, err := tree.Eval(rule.Params{
		"foo": "bar",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: bar
}

func ExampleTrue() {
	tree := rule.True()

	val, err := tree.Eval(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val.Data)
	// Output: true
}
