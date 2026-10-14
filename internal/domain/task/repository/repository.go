package repository

import (
	"context"
	"database/sql"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/pagination"
)

type TaskFilter struct {
	TeamID     string
	Status     *int
	OwnerID    string
	AssigneeID string
	Search     string
}

type TaskRepositoryInterface interface {
	// Transactions
	BeginTx(ctx context.Context) (*sql.Tx, error)
	Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error

	// Teams
	CreateTeam(ctx context.Context, tx *sql.Tx, team *model.Team) (*model.Team, error)
	FindTeamByID(ctx context.Context, teamID string) (*model.Team, error)
	FindUserTeams(ctx context.Context, userID string) ([]*model.Team, error)
	DeleteTeam(ctx context.Context, tx *sql.Tx, teamID string) error

	// Team Members
	AddTeamMember(ctx context.Context, tx *sql.Tx, member *model.TeamMember) error
	FindTeamMember(ctx context.Context, teamID, userID string) (*model.TeamMember, error)
	FindActiveTeamMembers(ctx context.Context, teamID string) ([]*model.TeamMember, error)
	SoftDeleteTeamMember(ctx context.Context, tx *sql.Tx, teamID, userID string) error

	// Tasks
	CreateTask(ctx context.Context, tx *sql.Tx, task *model.Task) (*model.Task, error)
	FindTaskByID(ctx context.Context, taskID string) (*model.Task, error)
	FindTasks(ctx context.Context, filter TaskFilter, p pagination.Request) ([]*model.Task, *pagination.Meta, error)
	UpdateTask(ctx context.Context, tx *sql.Tx, task *model.Task) error
	DeleteTask(ctx context.Context, tx *sql.Tx, taskID string) error

	// Collaborators & Logs
	AssignTaskCollaborator(ctx context.Context, tx *sql.Tx, col *model.TaskCollaborator) error
	CreateTaskLog(ctx context.Context, tx *sql.Tx, log *model.TaskLog) error
	FindTaskCollaborators(ctx context.Context, taskID string) ([]*model.TaskCollaborator, error)
}
