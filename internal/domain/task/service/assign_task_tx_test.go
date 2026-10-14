package service_test

import (
	"context"
	"errors"
	"testing"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService_AssignTask_Transaction(t *testing.T) {
	ctx := context.Background()

	t.Run("succeeds and executes collaborator assignment and task log in transaction", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		task := &model.Task{ID: "task-1", TeamID: "team-1", OwnerID: "usr-owner"}
		team := &model.Team{ID: "team-1", OwnerID: "usr-owner"}
		targetMember := &model.TeamMember{TeamID: "team-1", UserID: "usr-target"}

		repo.On("FindTaskByID", ctx, "task-1").Return(task, nil)
		repo.On("FindTeamByID", ctx, "team-1").Return(team, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-target").Return(targetMember, nil)
		repo.On("FindTaskCollaborators", ctx, "task-1").Return([]*model.TaskCollaborator{}, nil)

		repo.On("AssignTaskCollaborator", ctx, mock.Anything, mock.MatchedBy(func(col *model.TaskCollaborator) bool {
			return col.TaskID == "task-1" && col.UserID == "usr-target" && col.RoleType == model.RoleAssignee
		})).Return(nil)

		repo.On("CreateTaskLog", ctx, mock.Anything, mock.MatchedBy(func(log *model.TaskLog) bool {
			return log.TaskID == "task-1" && log.NewAssigneeID == "usr-target" && log.Action == "ASSIGN"
		})).Return(nil)

		err := svc.AssignTask(ctx, "usr-owner", service.AssignTaskRequest{
			TaskID:   "task-1",
			UserID:   "usr-target",
			RoleType: model.RoleAssignee,
		})

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("fails and triggers rollback when CreateTaskLog fails in transaction", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		task := &model.Task{ID: "task-1", TeamID: "team-1", OwnerID: "usr-owner"}
		team := &model.Team{ID: "team-1", OwnerID: "usr-owner"}
		targetMember := &model.TeamMember{TeamID: "team-1", UserID: "usr-target"}

		repo.On("FindTaskByID", ctx, "task-1").Return(task, nil)
		repo.On("FindTeamByID", ctx, "team-1").Return(team, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-target").Return(targetMember, nil)
		repo.On("FindTaskCollaborators", ctx, "task-1").Return([]*model.TaskCollaborator{}, nil)

		// AssignTaskCollaborator sukses di dalam tx
		repo.On("AssignTaskCollaborator", ctx, mock.Anything, mock.Anything).Return(nil)

		// CreateTaskLog GAGAL di dalam tx
		repo.On("CreateTaskLog", ctx, mock.Anything, mock.Anything).Return(errors.New("db error writing task log"))

		err := svc.AssignTask(ctx, "usr-owner", service.AssignTaskRequest{
			TaskID:   "task-1",
			UserID:   "usr-target",
			RoleType: model.RoleAssignee,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to assign collaborator")
		repo.AssertExpectations(t)
	})
}
