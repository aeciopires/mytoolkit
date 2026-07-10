// Package apperr defines the shared application error shape used by
// internal/tools, internal/httpapi, and internal/cli so error codes and
// HTTP status mapping are defined once instead of per feature.
package apperr

import "fmt"

// Error is a structured application error with a stable machine-readable
// code and the HTTP status it maps to.
type Error struct {
	Code    string
	Message string
	Status  int
}

func (e *Error) Error() string {
	return e.Message
}

// New creates an Error with the given HTTP status, code, and message.
func New(status int, code, message string) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

// Newf creates an Error with a formatted message.
func Newf(status int, code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Status: status}
}

// ErrEmptyInput is returned by tools that require non-empty input.
var ErrEmptyInput = New(400, "EMPTY_INPUT", "input must not be empty")

// OneOf validates that value is one of allowed, returning an Error otherwise.
// Used by every Mode/Algorithm-style enum across tools.
func OneOf[T comparable](field string, value T, allowed ...T) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return Newf(400, "INVALID_"+"OPTION", "%s must be one of %v, got %v", field, allowed, value)
}
