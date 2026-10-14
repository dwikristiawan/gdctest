package controller

import (
	"gdctest/pkg/echo"
)

func (ctrl *TaskController) GetTeamMembers(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	teamID := c.Param("id")
	members, err := ctrl.svc.GetTeamMembers(c.Request().Context(), teamID, userID)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, members, nil)
}
