package api

import (
	"github.com/heetch/regula/rule"
)

// Value is the response sent to the client after an eval.
type Value struct {
	Data string `json:"data"`
	Type string `json:"type"`
}

// Error is a generic error response.
type Error struct {
	Err string `json:"error"`
}

func (e Error) Error() string {
	return e.Err
}

// Ruleset holds a ruleset and its metadata.
type Ruleset struct {
	Path    string        `json:"path"`
	Ruleset *rule.Ruleset `json:"ruleset"`
}

// List of possible events executed against a ruleset.
const (
	PutEvent    = "PUT"
	DeleteEvent = "DELETE"
)

// Event describes an event occured on a ruleset.
type Event struct {
	Type    string        `json:"type"`
	Path    string        `json:"path"`
	Ruleset *rule.Ruleset `json:"ruleset"`
}
