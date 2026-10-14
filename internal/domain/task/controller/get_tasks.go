package controller

import (
	"strconv"
	"strings"

	"gdctest/internal/domain/task/repository"
	"gdctest/pkg/echo"
	"gdctest/pkg/pagination"
)

func (ctrl *TaskController) GetTasks(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	filter := repository.TaskFilter{
		TeamID:     strings.TrimSpace(c.QueryParam("team_id")),
		AssigneeID: strings.TrimSpace(c.QueryParam("assignee_id")),
	}

	search := strings.TrimSpace(c.QueryParam("search"))
	if search == "" {
		search = strings.TrimSpace(c.QueryParam("title"))
	}
	filter.Search = search

	if statusStr := strings.TrimSpace(c.QueryParam("status")); statusStr != "" {
		if st, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &st
		}
	}

	p := pagination.FromEcho(c)

	tasks, meta, err := ctrl.svc.GetTasks(c.Request().Context(), userID, filter, p)
	if err != nil {
		return echo.Respond(c, nil, err)
	}

	return echo.Respond(c, map[string]any{
		"items": tasks,
		"meta":  meta,
	}, nil)
}
