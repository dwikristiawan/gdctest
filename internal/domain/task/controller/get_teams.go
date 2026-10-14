package controller

import (
	"gdctest/pkg/echo"
)

func (ctrl *TaskController) GetTeams(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	teams, err := ctrl.svc.GetTeams(c.Request().Context(), userID)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, teams, nil)
}
