package controller

import (
	"net/http"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
	"gdctest/pkg/httputil"
)

type SetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

func (ctrl *AuthController) SetPassword(c echo.Context) error {
	var req SetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	ip := httputil.GetRealIP(c)
	user, err := ctrl.svc.SetPassword(c.Request().Context(), req.Token, req.Username, req.Password, ip)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, user, nil)
}
