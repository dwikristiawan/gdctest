package service

import (
	"context"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/repository"
	"gdctest/pkg/pagination"

	"github.com/redis/go-redis/v9"
)

type CreateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddMemberRequest struct {
	UserID   string `json:"user_id"`
	RoleType int    `json:"role_type"`
}

type CreateTaskRequest struct {
	TeamID         string `json:"team_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Status         int    `json:"status"`
	IdempotencyKey string `json:"-"`
}

type UpdateTaskRequest struct {
	ID          string  `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
}

type AssignTaskRequest struct {
	TaskID   string `json:"task_id"`
	UserID   string `json:"user_id"`
	RoleType int    `json:"role_type"`
}

type TaskServiceInterface interface {
	// Teams
	CreateTeam(ctx context.Context, ownerID string, req CreateTeamRequest) (*model.Team, error)
	GetTeams(ctx context.Context, userID string) ([]*model.Team, error)
	DeleteTeam(ctx context.Context, teamID, userID string) error

	// Team Members
	AddMember(ctx context.Context, teamID, currentUserID string, req AddMemberRequest) error
	GetTeamMembers(ctx context.Context, teamID, currentUserID string) ([]*model.TeamMember, error)
	RemoveMember(ctx context.Context, teamID, currentUserID, targetUserID string) error

	// Tasks
	CreateTask(ctx context.Context, currentUserID string, req CreateTaskRequest) (*model.Task, error)
	GetTasks(ctx context.Context, currentUserID string, filter repository.TaskFilter, p pagination.Request) ([]*model.Task, *pagination.Meta, error)
	GetTaskDetail(ctx context.Context, taskID, currentUserID string) (*model.Task, error)
	UpdateTask(ctx context.Context, currentUserID string, req UpdateTaskRequest) (*model.Task, error)
	DeleteTask(ctx context.Context, taskID, currentUserID string) error
	AssignTask(ctx context.Context, currentUserID string, req AssignTaskRequest) error
}

type TaskService struct {
	repo repository.TaskRepositoryInterface
	rdb  *redis.Client
}

func NewTaskService(repo repository.TaskRepositoryInterface, rdb ...*redis.Client) TaskServiceInterface {
	var client *redis.Client
	if len(rdb) > 0 {
		client = rdb[0]
	}
	return &TaskService{repo: repo, rdb: client}
}
