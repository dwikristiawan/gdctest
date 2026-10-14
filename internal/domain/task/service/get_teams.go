package service

import (
	"context"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) GetTeams(ctx context.Context, userID string) ([]*model.Team, error) {
	teams, err := s.repo.FindUserTeams(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("failed to fetch user teams", err)
	}
	if teams == nil {
		teams = []*model.Team{}
	}
	return teams, nil
}
