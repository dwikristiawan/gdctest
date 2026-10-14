package controller

import (
	"net/http"

	"gdctest/internal/domain/task/service"
	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
)

type TaskControllerInterface interface {
	// Teams
	CreateTeam(c echo.Context) error
	GetTeams(c echo.Context) error
	DeleteTeam(c echo.Context) error
	AddMember(c echo.Context) error
	GetTeamMembers(c echo.Context) error
	RemoveMember(c echo.Context) error

	// Tasks
	CreateTask(c echo.Context) error
	GetTasks(c echo.Context) error
	GetTaskDetail(c echo.Context) error
	UpdateTask(c echo.Context) error
	DeleteTask(c echo.Context) error
	AssignTask(c echo.Context) error
}

type TaskController struct {
	svc service.TaskServiceInterface
}

func NewTaskController(svc service.TaskServiceInterface) TaskControllerInterface {
	return &TaskController{svc: svc}
}

func getUserID(c echo.Context) (string, error) {
	val := c.Get("user_id")
	if uid, ok := val.(string); ok && uid != "" {
		return uid, nil
	}
	return "", apperror.New(http.StatusUnauthorized, "unauthorized")
}
