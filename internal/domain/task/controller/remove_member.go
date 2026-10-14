package controller

import (
	"gdctest/pkg/echo"
)

func (ctrl *TaskController) RemoveMember(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	teamID := c.Param("id")
	targetUserID := c.Param("user_id")

	err = ctrl.svc.RemoveMember(c.Request().Context(), teamID, userID, targetUserID)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{"message": "member removed successfully"}, nil)
}
