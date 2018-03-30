/*
Package rules provides a rules engine implementation.

Usage of this package revolves around the concept of rulesets.

A Ruleset can be represented as a list of rules that can be evaluated against a set of parameters given by a caller.
Each rule is evaluated in order and if one matches with the given parameters it returns a result and the evaluation stops.
All the rules of a ruleset always return the same type.

A Store is capable of serving rulesets on demand to the Engine which evaluates them to return a result.

While the Store is stateful and can hold rulesets in-memory, fetch them over the network or read them from a file,
the Engine is stateless and simply queries the Store for rulesets. The rulesets are then evaluated to a type safe result and
returned to the caller.

	engine := NewEngine(store)

	s := engine.GetString("path/to/string/ruleset/key", rule.Params{
		"user-id": 123,
		"email": "example@provider.com",
	})

	i := engine.GetInt64("path/to/int/ruleset/key", rule.Params{
		"user-id": 123,
		"email": "example@provider.com",
	})
*/
package rules
