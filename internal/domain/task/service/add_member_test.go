package service_test

import (
	"context"
	"testing"
	"time"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService_AddMember(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when user_id is empty or invalid role", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		err := svc.AddMember(ctx, "team-1", "usr-1", service.AddMemberRequest{UserID: ""})
		assert.Error(t, err)

		err = svc.AddMember(ctx, "team-1", "usr-1", service.AddMemberRequest{UserID: "usr-2", RoleType: 99})
		assert.Error(t, err)
	})

	t.Run("fails when team not found", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-999").Return(nil, nil)

		err := svc.AddMember(ctx, "team-999", "usr-1", service.AddMemberRequest{UserID: "usr-2", RoleType: model.RoleTeamMember})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
	})

	t.Run("fails when caller is not owner and not team admin", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-regular").Return(&model.TeamMember{
			TeamID:   "team-1",
			UserID:   "usr-regular",
			RoleType: model.RoleTeamMember, // Member, not Admin
		}, nil)

		err := svc.AddMember(ctx, "team-1", "usr-regular", service.AddMemberRequest{UserID: "usr-target", RoleType: model.RoleTeamMember})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Forbidden")
	})

	t.Run("fails when target user is already an active member", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-target").Return(&model.TeamMember{
			TeamID:    "team-1",
			UserID:    "usr-target",
			RoleType:  model.RoleTeamMember,
			DeletedAt: nil, // Active!
		}, nil)

		err := svc.AddMember(ctx, "team-1", "usr-owner", service.AddMemberRequest{UserID: "usr-target", RoleType: model.RoleTeamMember})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already an active member")
	})

	t.Run("succeeds when target user was previously soft deleted or new", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		now := time.Now()
		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-target").Return(&model.TeamMember{
			TeamID:    "team-1",
			UserID:    "usr-target",
			RoleType:  model.RoleTeamMember,
			DeletedAt: &now, // Soft deleted!
		}, nil)

		repo.On("AddTeamMember", ctx, mock.MatchedBy(func(m *model.TeamMember) bool {
			return m.TeamID == "team-1" && m.UserID == "usr-target" && m.RoleType == model.RoleTeamMember
		})).Return(nil)

		err := svc.AddMember(ctx, "team-1", "usr-owner", service.AddMemberRequest{UserID: "usr-target", RoleType: model.RoleTeamMember})
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
