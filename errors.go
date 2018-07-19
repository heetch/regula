package regula

import "errors"

var (
	// ErrRulesetNotFound must be returned when no ruleset is found for a given key.
	ErrRulesetNotFound = errors.New("ruleset not found")

	// ErrTypeMismatch is returned when the evaluated rule doesn't return the expected result type.
	ErrTypeMismatch = errors.New("type returned by rule doesn't match")

	// ErrParamTypeMismatch is returned when a parameter type is different from expected.
	ErrParamTypeMismatch = errors.New("parameter type mismatches")

	// ErrParamNotFound is returned when a parameter is not defined.
	ErrParamNotFound = errors.New("parameter not found")

	// ErrNoMatch is returned when the rule doesn't match the given context.
	ErrNoMatch = errors.New("rule doesn't match the given context")

	// ErrRulesetIncoherentType is returned when a ruleset contains rules of different types
	ErrRulesetIncoherentType = errors.New("types in ruleset are incoherent")
)
