package regula

import "errors"

var (
	// ErrTypeMismatch is returned when the evaluated rule doesn't return the expected result type.
	ErrTypeMismatch = errors.New("type returned by rule doesn't match")

	// ErrRulesetNotFound must be returned when no ruleset is found for a given key.
	ErrRulesetNotFound = errors.New("ruleset not found")

	// ErrRulesetIncoherentType is returned when a ruleset contains rules of different types.
	ErrRulesetIncoherentType = errors.New("types in ruleset are incoherent")

	// ErrBadRulesetName is returned if the name of the ruleset is badly formatted.
	ErrBadRulesetName = errors.New("bad ruleset name")
)
