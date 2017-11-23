package rule

import (
	"errors"
	"go/token"
	"strconv"
)

var (
	// ErrNoMatch is returned when the rule doesn't match the given context.
	ErrNoMatch = errors.New("rule doesn't match the given context")
)

// A Value is the result of a Node evaluation.
type Value struct {
	Type string
	Data string
}

// NewBoolValue creates a bool type value.
func NewBoolValue(value bool) *Value {
	return &Value{
		Type: "bool",
		Data: strconv.FormatBool(value),
	}
}

// NewStringValue creates a string type value.
func NewStringValue(value string) *Value {
	return &Value{
		Type: "string",
		Data: value,
	}
}

func (v *Value) compare(op token.Token, other *Value) bool {
	if op != token.EQL {
		return false
	}

	return *v == *other
}

// Equal reports whether v and other represent the same value.
func (v *Value) Equal(other *Value) bool {
	return v.compare(token.EQL, other)
}

// Eval evaluates the rule against the given context.
// If it matches it returns a result, otherwise it returns ErrNoMatch
// or any encountered error.
func (r *Rule) Eval(ctx map[string]string) (*Result, error) {
	value, err := r.Root.Eval(ctx)
	if err != nil {
		return nil, err
	}

	if value.Type != "bool" {
		return nil, errors.New("invalid rule returning non boolean value")
	}

	ok, err := strconv.ParseBool(value.Data)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrNoMatch
	}

	return r.Result, nil
}
