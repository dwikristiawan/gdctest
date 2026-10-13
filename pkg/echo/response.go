package echo

import (
	"errors"
	"net/http"

	"gdctest/pkg/apperror"

	echoV4 "github.com/labstack/echo/v4"
)

type APIResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// Respond is the unified 1-line responder for all controllers.
func Respond(c echoV4.Context, data any, err error) error {
	if err != nil {
		status := http.StatusInternalServerError

		var appErr *apperror.Error
		if errors.As(err, &appErr) && appErr.Code != 0 {
			status = appErr.Code
		}

		return c.JSON(status, APIResponse{
			Message: err.Error(),
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, APIResponse{
		Message: "Successful",
		Data:    data,
	})
}
