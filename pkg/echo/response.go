package echo

import (
	"errors"
	"net/http"

	"time"

	"gdctest/pkg/apperror"
	"gdctest/pkg/logger"

	echoV4 "github.com/labstack/echo/v4"
)

type APIResponse struct {
	Status    bool   `json:"status"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Error     any    `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

func httpStatusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	default:
		if status >= 500 {
			return "INTERNAL_SERVER_ERROR"
		}
		return "CLIENT_ERROR"
	}
}

func extractRealError(err error) string {
	if err == nil {
		return ""
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		if appErr.Err != nil {
			if appErr.Detail != "" {
				return appErr.Detail + ": " + appErr.Err.Error()
			}
			return appErr.Err.Error()
		}
		if appErr.Detail != "" {
			return appErr.Detail
		}
	}
	return err.Error()
}

// Respond is the unified 1-line responder for all controllers.
func Respond(c echoV4.Context, data any, err error) error {
	if err != nil {
		status := http.StatusInternalServerError
		respMessage := "Internal Server Error"

		var appErr *apperror.Error
		if errors.As(err, &appErr) && appErr.Code != 0 {
			status = appErr.Code
		}

		errorPayload := respMessage
		if status >= http.StatusInternalServerError {
			realErrMsg := extractRealError(err)
			if c != nil {
				c.Set("real_error", realErrMsg)
				if c.Request() != nil {
					logger.FromContext(c.Request().Context()).
						WithField("real_error", realErrMsg).
						WithError(err).
						Error("internal server error")
				}
			}
			respMessage = apperror.Internal("General error").Error()
			errorPayload = "General error"
		} else {
			respMessage = err.Error()
			if appErr != nil && appErr.Detail != "" {
				errorPayload = appErr.Detail
			}
		}

		now := time.Now().UTC().Format(time.RFC3339)
		return c.JSON(status, APIResponse{
			Status:    false,
			Code:      httpStatusToCode(status),
			Message:   respMessage,
			Error:     errorPayload,
			Timestamp: now,
		})
	}

	now := time.Now().UTC().Format(time.RFC3339)
	return c.JSON(http.StatusOK, APIResponse{
		Status:    true,
		Message:   "Successful",
		Data:      data,
		Timestamp: now,
	})
}

// RespondWithStatus allows specifying custom HTTP status code on success (e.g. 201 Created).
func RespondWithStatus(c echoV4.Context, successStatus int, data any, err error) error {
	if err != nil {
		return Respond(c, data, err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	return c.JSON(successStatus, APIResponse{
		Status:    true,
		Message:   "Successful",
		Data:      data,
		Timestamp: now,
	})
}
