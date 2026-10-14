package service_test

import (
	"context"
	"errors"
	"testing"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"

	"github.com/stretchr/testify/assert"
)

func TestTaskService_DeleteTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when team does not exist", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-999").Return(nil, nil)

		err := svc.DeleteTeam(ctx, "team-999", "usr-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
	})

	t.Run("fails when user is not the team owner", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{
			ID:      "team-1",
			OwnerID: "usr-owner",
		}, nil)

		err := svc.DeleteTeam(ctx, "team-1", "usr-intruder")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Forbidden")
	})

	t.Run("fails when repo delete fails", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{
			ID:      "team-1",
			OwnerID: "usr-owner",
		}, nil)
		repo.On("DeleteTeam", ctx, "team-1").Return(errors.New("db error"))

		err := svc.DeleteTeam(ctx, "team-1", "usr-owner")
		assert.Error(t, err)
	})

	t.Run("succeeds when user is team owner", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{
			ID:      "team-1",
			OwnerID: "usr-owner",
		}, nil)
		repo.On("DeleteTeam", ctx, "team-1").Return(nil)

		err := svc.DeleteTeam(ctx, "team-1", "usr-owner")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
