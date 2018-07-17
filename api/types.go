package api

import (
	"fmt"
	"net/http"

	"github.com/heetch/regula/rule"
)

// Value is the response sent to the client after an eval.
type Value struct {
	Data    string `json:"data"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

// Error is a generic error response.
type Error struct {
	Err string `json:"error"`
	// Used by clients to return the origin server response
	Response *http.Response `json:"-"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		e.Response.Request.Method,
		e.Response.Request.URL,
		e.Response.StatusCode,
		e.Err)
}

// Ruleset holds a ruleset and its metadata.
type Ruleset struct {
	Path    string        `json:"path"`
	Version string        `json:"version"`
	Ruleset *rule.Ruleset `json:"ruleset"`
}

// Rulesets holds a list of rulesets.
type Rulesets struct {
	Rulesets []Ruleset `json:"rulesets"`
	Revision string    `json:"revision"`
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
	Version string        `json:"version"`
	Ruleset *rule.Ruleset `json:"ruleset"`
}

// Events holds a list of events occured on a group of rulesets.
type Events struct {
	Events   []Event `json:"events"`
	Revision string  `json:"revision"`
}
