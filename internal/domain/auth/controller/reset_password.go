package controller

import (
	"net/http"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
	"gdctest/pkg/httputil"
)

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

func (ctrl *AuthController) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	ip := httputil.GetRealIP(c)
	err := ctrl.svc.ResetPassword(c.Request().Context(), req.Token, req.Password, ip)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{
		"message": "password reset successfully, all active sessions cleared",
	}, nil)
}
