package controller

import (
	"net/http"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
	"gdctest/pkg/httputil"
)

type RegisterRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (ctrl *AuthController) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	ip := httputil.GetRealIP(c)
	err := ctrl.svc.Register(c.Request().Context(), req.Email, ip)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{
		"message": "verification email sent, please check your inbox (valid 30 minutes)",
	}, nil)
}
