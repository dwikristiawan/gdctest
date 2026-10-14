package controller

import (
	"net/http"

	"gdctest/internal/domain/task/service"
	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
)

type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

func (ctrl *TaskController) CreateTeam(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	var req CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	team, err := ctrl.svc.CreateTeam(c.Request().Context(), userID, service.CreateTeamRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, team, nil)
}
