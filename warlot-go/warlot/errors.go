package warlot

import "fmt"

// APIError represents a non-success HTTP response from the API.
type APIError struct {
	StatusCode int
	Body       string
	Message    string
	Code       string      // Optional server-provided code.
	Details    interface{} // Optional server-provided details.
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.Body
	}
	if e.Code != "" {
		return fmt.Sprintf("warlot API %d (%s): %s", e.StatusCode, e.Code, msg)
	}
	return fmt.Sprintf("warlot API %d: %s", e.StatusCode, msg)
}
