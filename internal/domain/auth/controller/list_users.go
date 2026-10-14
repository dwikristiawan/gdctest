package controller

import (
	"gdctest/pkg/echo"
	"gdctest/pkg/pagination"
)

// ListUsers handles GET /v1.0/auth/users
func (ctrl *AuthController) ListUsers(c echo.Context) error {
	p := pagination.FromEcho(c)

	users, meta, err := ctrl.svc.ListUsers(c.Request().Context(), p)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]any{
		"items": users,
		"meta":  meta,
	}, nil)
}
