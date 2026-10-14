package controller

import (
	"net/http"

	"gdctest/internal/domain/task/service"
	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
)

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
}

func (ctrl *TaskController) UpdateTask(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	taskID := c.Param("id")

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	task, err := ctrl.svc.UpdateTask(c.Request().Context(), userID, service.UpdateTaskRequest{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, task, nil)
}
