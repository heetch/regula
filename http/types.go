package http

import (
	"fmt"
	"net/http"
)

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
