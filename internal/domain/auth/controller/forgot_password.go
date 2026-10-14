package controller

import (
	"net/http"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
	"gdctest/pkg/httputil"
)

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (ctrl *AuthController) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	ip := httputil.GetRealIP(c)
	err := ctrl.svc.ForgotPassword(c.Request().Context(), req.Email, ip)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{
		"message": "reset password email sent (valid 15 minutes)",
	}, nil)
}
