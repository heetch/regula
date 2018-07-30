package rule

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/tidwall/gjson"
)

var (
	// ErrRulesetNotFound must be returned when no ruleset is found for a given key.
	ErrRulesetNotFound = errors.New("ruleset not found")

	// ErrParamTypeMismatch is returned when a parameter type is different from expected.
	ErrParamTypeMismatch = errors.New("parameter type mismatches")

	// ErrParamNotFound is returned when a parameter is not defined.
	ErrParamNotFound = errors.New("parameter not found")

	// ErrNoMatch is returned when the rule doesn't match the given params.
	ErrNoMatch = errors.New("rule doesn't match the given params")

	// ErrRulesetIncoherentType is returned when a ruleset contains rules of different types.
	ErrRulesetIncoherentType = errors.New("types in ruleset are incoherent")

	// ErrParameterBadFormat is returned if the name of the parameter is badly formatted.
	ErrParameterBadFormat = errors.New("bad parameter name")
)

// A Rule represents a logical boolean expression that evaluates to a result.
type Rule struct {
	Expr   Expr
	Result *Value
}

// New creates a rule with the given expression and that returns the given result on evaluation.
func New(expr Expr, result *Value) *Rule {
	return &Rule{
		Expr:   expr,
		Result: result,
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Rule) UnmarshalJSON(data []byte) error {
	tree := struct {
		Expr   json.RawMessage
		Result *Value
	}{}

	err := json.Unmarshal(data, &tree)
	if err != nil {
		return err
	}

	if tree.Result.Type == "" {
		return errors.New("invalid rule result type")
	}

	res := gjson.Get(string(tree.Expr), "kind")
	n, err := unmarshalExpr(res.Str, []byte(tree.Expr))
	if err != nil {
		return err
	}

	r.Expr = n
	r.Result = tree.Result
	return err
}

// Eval evaluates the rule against the given params.
// If it matches it returns a result, otherwise it returns ErrNoMatch
// or any encountered error.
func (r *Rule) Eval(params Params) (*Value, error) {
	value, err := r.Expr.Eval(params)
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
