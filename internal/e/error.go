package errors

import (
	"fmt"
	"net/http"
)

// Error defines the interface for custom errors.
type Error interface {
	Code() int
	Detail() string
}

type httpError struct {
	detail string
	code   int
}

func (e httpError) Error() string {
	return fmt.Sprintf("code: %d, detail: '%s'", e.code, e.detail)
}

func (e httpError) Code() int {
	return e.code
}

func (e httpError) Detail() string {
	return e.detail
}

// NewInternal creates a new internal server error.
func NewInternal(detail string) Error {
	return httpError{
		detail: detail,
		code:   http.StatusInternalServerError,
	}
}

// NewInternalf creates a new internal server error with a formatted detail message.
func NewInternalf(template string, args ...interface{}) Error {
	return httpError{
		detail: fmt.Sprintf(template, args...),
		code:   http.StatusInternalServerError,
	}
}

// NewNotFound creates a new not found error.
func NewNotFound(detail string) Error {
	return httpError{
		detail: detail,
		code:   http.StatusNotFound,
	}
}

// NewBadRequest creates a new bad request error.
func NewBadRequest(detail string) Error {
	return httpError{
		detail: detail,
		code:   http.StatusBadRequest,
	}
}
