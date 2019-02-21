package store

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/heetch/regula"
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

		data, err := json.Marshal(sig.ParamTypes)
		if err != nil {
			return errors.Wrap(err, "failed to encode param types")
		}

		return &ValidationError{
			Field:  "param",
			Value:  string(data),
			Reason: err.Error(),
		}
	}

	for name := range sig.ParamTypes {
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
	if !rgxRuleset.MatchString(path) {
		return &ValidationError{
			Field:  "path",
			Value:  path,
			Reason: "invalid format",
		}
	}

	return nil
}
