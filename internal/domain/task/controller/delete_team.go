package controller

import (
	"gdctest/pkg/echo"
)

func (ctrl *TaskController) DeleteTeam(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	teamID := c.Param("id")
	if err := ctrl.svc.DeleteTeam(c.Request().Context(), teamID, userID); err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{"message": "team deleted successfully"}, nil)
}
