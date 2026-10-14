package controller

import (
	"net/http"
	"strings"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
)

func (ctrl *AuthController) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return echo.Respond(c, nil, apperror.New(http.StatusUnauthorized, "missing or invalid authorization header"))
	}

	err := ctrl.svc.Logout(c.Request().Context(), parts[1])
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{"message": "logged out successfully"}, nil)
}
