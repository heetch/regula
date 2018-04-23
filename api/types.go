package api

import (
	"github.com/heetch/rules-engine/rule"
)

// Error is a generic error response.
type Error struct {
	Err string `json:"error"`
}

func (e Error) Error() string {
	return e.Err
}

// Ruleset holds a ruleset and its metadata.
type Ruleset struct {
	Name    string        `json:"name"`
	Ruleset *rule.Ruleset `json:"ruleset"`
}
