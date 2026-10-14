package service_test

import (
	"context"
	"testing"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"

	"github.com/stretchr/testify/assert"
)

func TestTaskService_GetTeamMembers(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when team does not exist", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-999").Return(nil, nil)

		res, err := svc.GetTeamMembers(ctx, "team-999", "usr-1")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "Not Found")
	})

	t.Run("fails when caller is not owner and not team member", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-stranger").Return(nil, nil)

		res, err := svc.GetTeamMembers(ctx, "team-1", "usr-stranger")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "Forbidden")
	})

	t.Run("succeeds when caller is member", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindTeamByID", ctx, "team-1").Return(&model.Team{ID: "team-1", OwnerID: "usr-owner"}, nil)
		repo.On("FindTeamMember", ctx, "team-1", "usr-member").Return(&model.TeamMember{
			TeamID: "team-1",
			UserID: "usr-member",
		}, nil)

		members := []*model.TeamMember{
			{TeamID: "team-1", UserID: "usr-owner", RoleType: model.RoleTeamAdmin},
			{TeamID: "team-1", UserID: "usr-member", RoleType: model.RoleTeamMember},
		}
		repo.On("FindActiveTeamMembers", ctx, "team-1").Return(members, nil)

		res, err := svc.GetTeamMembers(ctx, "team-1", "usr-member")
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})
}
