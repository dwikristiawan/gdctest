package controller

import (
	"gdctest/pkg/echo"
)

func (ctrl *TaskController) DeleteTask(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	taskID := c.Param("id")
	if err := ctrl.svc.DeleteTask(c.Request().Context(), taskID, userID); err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{"message": "task deleted successfully"}, nil)
}
