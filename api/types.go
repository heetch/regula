package api

import (
	"fmt"
	"net/http"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
)

// EvalResult is the response sent to the client after an eval.
type EvalResult struct {
	Value   *rule.Value `json:"value"`
	Version string      `json:"version"`
}

// Error is a generic error response.
type Error struct {
	Err      string         `json:"error"`
	Response *http.Response `json:"-"` // Used by clients to return the original server response
}

func (e Error) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		e.Response.Request.Method,
		e.Response.Request.URL,
		e.Response.StatusCode,
		e.Err)
}

// Rulesets holds a list of rulesets.
type Rulesets struct {
	Rulesets []regula.Ruleset `json:"rulesets"`
	Revision string           `json:"revision"`
	Continue string           `json:"continue,omitempty"`
}

// List of possible events executed against a ruleset.
const (
	PutEvent = "PUT"
)

// Event describes an event occured on a ruleset.
type Event struct {
	Type    string       `json:"type"`
	Path    string       `json:"path"`
	Version string       `json:"version"`
	Rules   []*rule.Rule `json:"rules"`
}

// Events holds a list of events occured on a group of rulesets.
type Events struct {
	Events   []Event `json:"events,omitempty"`
	Revision string  `json:"revision,omitempty"`
	Timeout  bool    `json:"timeout,omitempty"`
}
