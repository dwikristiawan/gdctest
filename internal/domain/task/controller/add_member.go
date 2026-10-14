package controller

import (
	"net/http"

	"gdctest/internal/domain/task/service"
	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
)

type AddMemberRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	RoleType int    `json:"role_type"`
}

func (ctrl *TaskController) AddMember(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	teamID := c.Param("id")

	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return echo.Respond(c, nil, apperror.New(http.StatusBadRequest, err.Error()))
	}

	err = ctrl.svc.AddMember(c.Request().Context(), teamID, userID, service.AddMemberRequest{
		UserID:   req.UserID,
		RoleType: req.RoleType,
	})
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]string{"message": "member added successfully"}, nil)
}
