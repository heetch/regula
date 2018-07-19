/*
Package regula provides a rules engine implementation.

Usage of this package revolves around the concept of rulesets.

A ruleset can be represented as a list of rules that can be evaluated against a set of parameters given by a caller.
Each rule is evaluated in order and if one matches with the given parameters it returns a result and the evaluation stops.
All the rules of a ruleset always return the same type.

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

To query and evaluate rulesets with a set of parameters, the engine must be used.
An engine takes a getter which is responsible of serving rulesets on demand, the engine then evaluates the returned ruleset to return a result.

While the getter is stateful and can hold rulesets in-memory, fetch them over the network or read them from a file,
the engine is stateless and simply queries the getter for rulesets. The rulesets are then evaluated to a type safe result and
returned to the caller.

	engine := NewEngine(getter)

	s, err := engine.GetString("path/to/string/ruleset/key", rule.Params{
		"user-id": 123,
		"email": "example@provider.com",
	})

	i, err := engine.GetInt64("path/to/int/ruleset/key", rule.Params{
		"user-id": 123,
		"email": "example@provider.com",
	})
*/
package regula
