package service_test

import (
	"context"
	"errors"
	"testing"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"

	"github.com/stretchr/testify/assert"
)

func TestTaskService_GetTeams(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repo fails", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		repo.On("FindUserTeams", ctx, "usr-1").Return(nil, errors.New("db error"))

		teams, err := svc.GetTeams(ctx, "usr-1")
		assert.Error(t, err)
		assert.Nil(t, teams)
	})

	t.Run("returns list of teams", func(t *testing.T) {
		repo := new(MockTaskRepository)
		svc := service.NewTaskService(repo)

		expected := []*model.Team{
			{ID: "team-1", Name: "Team Alpha"},
		}
		repo.On("FindUserTeams", ctx, "usr-1").Return(expected, nil)

		teams, err := svc.GetTeams(ctx, "usr-1")
		assert.NoError(t, err)
		assert.Len(t, teams, 1)
		assert.Equal(t, "Team Alpha", teams[0].Name)
	})
}
