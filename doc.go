/*
Package regula provides a rules engine implementation.

Usage of this package revolves around the concept of rulesets.

A ruleset can be represented as a list of rules that can be evaluated against a set of parameters given by a caller.
Each rule is evaluated in order and if one matches with the given parameters it returns a result and the evaluation stops.
All the rules of a ruleset always return the same type.

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

To query and evaluate rulesets with a set of parameters, the engine must be used.
An engine takes an evaluator which is responsible of evaluating rulesets on demand and return a value, the engine then parses the value into a type safe result
and return it to the caller.

While the evaluator is stateful and can hold rulesets in-memory, fetch them over the network or read them from a file,
the engine is stateless and simply deleguates the evaluation to the evaluator.

	engine := regula.NewEngine(evaluator)

	s, res, err := engine.GetString("path/to/string/ruleset/key", regula.Params{
		"user-id": 123,
		"email": "example@provider.com",
	})

	i, res, err := engine.GetInt64("path/to/int/ruleset/key", regula.Params{
		"user-id": 123,
		"email": "example@provider.com",
	})
*/
package regula
