package regula

import (
	"fmt"
)

// Error is returned if two signatures are not equal.
type Error struct {
	Field  string
	Value  string
	Reason string
}

// Error makes the Error type to be compliant with the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("invalid %s with value '%s': %s", e.Field, e.Value, e.Reason)
}

// Signature represents the signature of a ruleset.
type Signature struct {
	ReturnType string
	ParamTypes map[string]string
}

// NewSignature returns the Signature of the given ruleset.
func NewSignature(rs *Ruleset) *Signature {
	pt := make(map[string]string)
	for _, p := range rs.Params() {
		pt[p.Name] = p.Type
	}

	return &Signature{
		ParamTypes: pt,
		ReturnType: rs.Type,
	}
}

// Equal checks if the given signature matches the current one.
func (s *Signature) Equal(other *Signature) (bool, error) {
	if s.ReturnType != other.ReturnType {
		return false, &Error{
			Field:  "return type",
			Value:  other.ReturnType,
			Reason: fmt.Sprintf("signature mismatch: return type must be of type %s", s.ReturnType),
		}
	}

	for name, tp := range other.ParamTypes {
		stp, ok := s.ParamTypes[name]
		if !ok {
			return false, &Error{
				Field:  "param",
				Value:  name,
				Reason: "signature mismatch: unknown parameter",
			}
		}

		if tp != stp {
			return false, &Error{
				Field:  "param type",
				Value:  tp,
				Reason: fmt.Sprintf("signature mismatch: param must be of type %s", stp),
			}
		}
	}

	return true, nil
}
