package service_test

import (
	"context"
	"testing"
	"time"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"

	"github.com/stretchr/testify/assert"
)

func TestTaskService_RemoveMember(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when target user_id is empty", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		err := svc.RemoveMember(ctx, "team-1", "usr-1", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("fails when team does not exist", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-999").Return(nil, nil)

		err := svc.RemoveMember(ctx, "team-999", "usr-1", "usr-2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
	})

	t.Run("fails when attempting to remove team owner", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)

		err := svc.RemoveMember(ctx, "team-1", "usr-admin", "usr-owner")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove team owner")
	})

	t.Run("fails when caller is unauthorized (not self, not owner, not admin)", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-member").Return(&model.TeamMember{
			TeamID:   "team-1",
			UserID:   "usr-member",
			RoleType: model.RoleTeamMember, // regular member
		}, nil)

		err := svc.RemoveMember(ctx, "team-1", "usr-member", "usr-other")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Forbidden")
	})

	t.Run("fails when target member is already removed or not found", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		now := time.Now()
		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-target").Return(&model.TeamMember{
			TeamID:    "team-1",
			UserID:    "usr-target",
			DeletedAt: &now,
		}, nil)

		err := svc.RemoveMember(ctx, "team-1", "usr-owner", "usr-target")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
	})

	t.Run("succeeds when team owner removes active member", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-target").Return(&model.TeamMember{
			TeamID:   "team-1",
			UserID:   "usr-target",
			RoleType: model.RoleTeamMember,
		}, nil)
		repo.On("SoftDeleteTeamMember", ctx, "team-1", "usr-target").Return(nil)

		err := svc.RemoveMember(ctx, "team-1", "usr-owner", "usr-target")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("succeeds when user removes themselves", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-self").Return(&model.TeamMember{
			TeamID: "team-1",
			UserID: "usr-self",
		}, nil)
		repo.On("SoftDeleteTeamMember", ctx, "team-1", "usr-self").Return(nil)

		err := svc.RemoveMember(ctx, "team-1", "usr-self", "usr-self")
		assert.NoError(t, err)
	})
}
