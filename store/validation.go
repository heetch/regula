package store

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

var (
	// regex used to validate ruleset names.
	rgxRuleset = regexp.MustCompile(`^[a-z]+(?:[a-z0-9-\/]?[a-z0-9])*$`)

	// regex used to validate parameters name.
	rgxParam = regexp.MustCompile(`^[a-z]+(?:[a-z0-9-]?[a-z0-9])*$`)

	// list of reserved words that shouldn't be used as parameters.
	reservedWords = []string{
		"version",
		"list",
		"eval",
		"watch",
		"revision",
	}
)

// ValidationError gives informations about the reason of failed validation.
type ValidationError struct {
	Field  string
	Value  string
	Reason string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s with value '%s': %s", v.Field, v.Value, v.Reason)
}

// IsValidationError indicates if the given error is a ValidationError pointer.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// ValidateSignature verifies if the return type and params names are valid.
func ValidateSignature(sig *regula.Signature) error {
	if err := sig.Validate(); err != nil {
		if err.Error() == "unsupported return type" {
			return &ValidationError{
				Field:  "returnType",
				Value:  sig.ReturnType,
				Reason: err.Error(),
			}
		}

		data, err := json.Marshal(sig.Params)
		if err != nil {
			return errors.Wrap(err, "failed to encode param types")
		}

		return &ValidationError{
			Field:  "param",
			Value:  string(data),
			Reason: err.Error(),
		}
	}

	for name := range sig.Params {
		err := ValidateParamName(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateParamName verifies if the given name matches the param name regex or is
// a reserved word.
// If not it returns a ValidationError.
func ValidateParamName(name string) error {
	if !rgxParam.MatchString(name) {
		return &ValidationError{
			Field:  "param",
			Value:  name,
			Reason: "invalid format",
		}
	}

	for _, w := range reservedWords {
		if name == w {
			return &ValidationError{
				Field:  "param",
				Value:  name,
				Reason: "reserved word",
			}
		}
	}

	return nil
}

// ValidatePath verifies if the given path matches the path regex.
func ValidatePath(path string) error {
	if path == "" || !rgxRuleset.MatchString(path) {
		return &ValidationError{
			Field:  "path",
			Value:  path,
			Reason: "invalid format",
		}
	}

	return nil
}

// ValidateRule verifies that the given rule respects the signature and contains valid data.
func ValidateRule(signature *regula.Signature, r *rule.Rule) error {
	for _, p := range r.Params() {
		if p.Kind != "param" {
			return &ValidationError{
				Field:  "param.kind",
				Value:  p.Kind,
				Reason: "param kind must be equal to 'param'",
			}
		}
		typ, ok := signature.Params[p.Name]
		if !ok {
			return &ValidationError{
				Field:  "param.name",
				Value:  p.Name,
				Reason: "unknown parameter",
			}
		}

		if p.Type != typ {
			return &ValidationError{
				Field:  "param.type",
				Value:  p.Type,
				Reason: fmt.Sprintf("param type must be '%s'", typ),
			}
		}
	}

	if r.Result == nil {
		return &ValidationError{
			Field:  "result",
			Value:  "null",
			Reason: "result is required",
		}
	}

	if r.Expr == nil {
		return &ValidationError{
			Field:  "expr",
			Value:  "null",
			Reason: "expr is required",
		}
	}

	if r.Result.Kind != "value" {
		return &ValidationError{
			Field:  "result.kind",
			Value:  r.Result.Kind,
			Reason: "result kind must be 'value'",
		}
	}

	if r.Result.Type != signature.ReturnType {
		return &ValidationError{
			Field:  "result.type",
			Value:  r.Result.Type,
			Reason: fmt.Sprintf("result type must be '%s'", signature.ReturnType),
		}
	}

	return ValidateValueData(r.Result)
}

// ValidateValueData verifies if the value data matches the value type.
func ValidateValueData(v *rule.Value) error {
	switch v.Type {
	case "string":
		return nil
	case "int64":
		_, err := strconv.ParseInt(v.Data, 10, 64)
		return err
	case "float64":
		_, err := strconv.ParseFloat(v.Data, 64)
		return err
	case "bool":
		_, err := strconv.ParseBool(v.Data)
		return err
	}

	return errors.New("unsupported param type")
}
