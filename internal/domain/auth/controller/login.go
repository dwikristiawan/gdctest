package controller

import (
	"net/http"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
	"gdctest/pkg/httputil"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (ctrl *AuthController) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	ip := httputil.GetRealIP(c)
	token, user, err := ctrl.svc.Login(c.Request().Context(), req.Email, req.Password, ip)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]any{
		"token": token,
		"user":  user,
	}, nil)
}
