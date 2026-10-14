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

func TestTaskService_CreateTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when name is empty", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		res, err := svc.CreateTeam(ctx, "usr-1", service.CreateTeamRequest{Name: "  "})
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "team name is required")
	})

	t.Run("fails when repo create team returns error", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("CreateTeam", ctx, mock.MatchedBy(func(tm *model.Team) bool {
			return tm.Name == "Engineering" && tm.OwnerID == "usr-1"
		})).Return(nil, errors.New("db error"))

		res, err := svc.CreateTeam(ctx, "usr-1", service.CreateTeamRequest{Name: "Engineering"})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("succeeds and automatically adds owner as admin member", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		expectedTeam := &model.Team{
			ID:      "team-1",
			Name:    "Engineering",
			OwnerID: "usr-1",
		}

		repo.On("CreateTeam", ctx, mock.MatchedBy(func(tm *model.Team) bool {
			return tm.Name == "Engineering" && tm.OwnerID == "usr-1"
		})).Return(expectedTeam, nil)

		repo.On("AddTeamMember", ctx, mock.MatchedBy(func(m *model.TeamMember) bool {
			return m.TeamID == "team-1" && m.UserID == "usr-1" && m.RoleType == model.RoleTeamAdmin
		})).Return(nil)

		res, err := svc.CreateTeam(ctx, "usr-1", service.CreateTeamRequest{Name: "Engineering"})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "team-1", res.ID)
		repo.AssertExpectations(t)
	})
}
