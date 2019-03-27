package errors

import "errors"

var (
	// ErrRulesetNotFound must be returned when no ruleset is found for a given key.
	ErrRulesetNotFound = errors.New("ruleset not found")

	// ErrRulesetVersionNotFound must be returned when a version of a ruleset is not found for a given key.
	ErrRulesetVersionNotFound = errors.New("ruleset version not found")

	// ErrParamTypeMismatch is returned when a parameter type is different from expected.
	ErrParamTypeMismatch = errors.New("parameter type mismatches")

	// ErrSignatureMismatch is returned when a rule doesn't respect a given signature.
	ErrSignatureMismatch = errors.New("signature mismatch")

	// ErrParamNotFound is returned when a parameter is not defined.
	ErrParamNotFound = errors.New("parameter not found")

	// ErrNoMatch is returned when the rule doesn't match the given params.
	ErrNoMatch = errors.New("rule doesn't match the given params")
)
