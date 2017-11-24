package rule_test

import (
	"fmt"
	"log"

	"github.com/heetch/rules-engine/rule"
)

func Example() {
	r := rule.New(
		rule.Eq(
			rule.ValueStr("foo"),
			rule.ParamStr("bar"),
		),
		rule.ReturnsStr("matched"),
	)

	ret, err := r.Eval(rule.Params{
		"bar": "foo",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ret.Value)
	// Output
	// matched
}
