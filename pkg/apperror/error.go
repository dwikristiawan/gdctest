package apperror

import (
	"fmt"
	"net/http"
)

// Error represents an application error with HTTP status and optional details.
// Format: "<StatusText>" or "<StatusText>. <Detail>" (e.g. "Bad Request. Email already registered")
type Error struct {
	Code   int    // HTTP status code (e.g. 400, 404, 500)
	Detail string // Additional contextual message
	Err    error  // Original internal error for root-cause tracking
}

func (e *Error) Error() string {
	statusText := http.StatusText(e.Code)
	if statusText == "" {
		statusText = "Unknown Error"
	}

	if e.Detail != "" {
		return fmt.Sprintf("%s. %s", statusText, e.Detail)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s. %s", statusText, e.Err.Error())
	}
	return statusText
}

func (e *Error) Unwrap() error {
	return e.Err
}

// New creates an Error where the message is "<StatusText>. <detail>" (or just StatusText if detail is empty).
func New(code int, detail string, err ...error) *Error {
	var originalErr error
	if len(err) > 0 {
		originalErr = err[0]
	}
	return &Error{
		Code:   code,
		Detail: detail,
		Err:    originalErr,
	}
}

func BadRequest(detail string, err ...error) *Error {
	return New(http.StatusBadRequest, detail, err...)
}

func NotFound(detail string, err ...error) *Error {
	return New(http.StatusNotFound, detail, err...)
}

func Unauthorized(detail string, err ...error) *Error {
	return New(http.StatusUnauthorized, detail, err...)
}

func Internal(detail string, err ...error) *Error {
	return New(http.StatusInternalServerError, detail, err...)
}
