package controller

import (
	"gdctest/pkg/echo"
)

func (ctrl *TaskController) GetTaskDetail(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	taskID := c.Param("id")
	task, err := ctrl.svc.GetTaskDetail(c.Request().Context(), taskID, userID)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, task, nil)
}
