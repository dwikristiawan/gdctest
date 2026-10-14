package controller

import (
	"net/http"
	"strings"

	"gdctest/internal/domain/task/service"
	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
)

type CreateTaskRequest struct {
	TeamID      string `json:"team_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

func (ctrl *TaskController) CreateTask(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	idempotencyKey := strings.TrimSpace(c.Request().Header.Get("Idempotency-Key"))

	task, err := ctrl.svc.CreateTask(c.Request().Context(), userID, service.CreateTaskRequest{
		TeamID:         req.TeamID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.RespondWithStatus(c, http.StatusCreated, task, nil)
}
