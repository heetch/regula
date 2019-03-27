package errors

// Error is used to define constant, comparable errors.
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	// ErrRulesetNotFound must be returned when no ruleset is found for a given key.
	ErrRulesetNotFound = Error("ruleset not found")

	// ErrRulesetVersionNotFound must be returned when a version of a ruleset is not found for a given key.
	ErrRulesetVersionNotFound = Error("ruleset version not found")

	// ErrParamTypeMismatch is returned when a parameter type is different from expected.
	ErrParamTypeMismatch = Error("parameter type mismatches")

	// ErrSignatureMismatch is returned when a rule doesn't respect a given signature.
	ErrSignatureMismatch = Error("signature mismatch")

	// ErrParamNotFound is returned when a parameter is not defined.
	ErrParamNotFound = Error("parameter not found")

	// ErrNoMatch is returned when the rule doesn't match the given params.
	ErrNoMatch = Error("rule doesn't match the given params")
)
