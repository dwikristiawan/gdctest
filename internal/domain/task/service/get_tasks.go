package service

import (
	"context"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/repository"
	"gdctest/pkg/apperror"
	"gdctest/pkg/pagination"
)

func (s *TaskService) GetTasks(
	ctx context.Context,
	currentUserID string,
	filter repository.TaskFilter,
	p pagination.Request,
) ([]*model.Task, *pagination.Meta, error) {
	if filter.TeamID != "" {
		team, err := s.repo.FindTeamByID(ctx, filter.TeamID)
		if err != nil {
			return nil, nil, apperror.Internal("failed to fetch team", err)
		}
		if team == nil {
			return nil, nil, apperror.NotFound("team not found")
		}

		if team.OwnerID != currentUserID {
			member, err := s.repo.FindTeamMember(ctx, filter.TeamID, currentUserID)
			if err != nil {
				return nil, nil, apperror.Internal("failed to verify membership", err)
			}
			if member == nil || member.DeletedAt != nil {
				return nil, nil, apperror.Forbidden("only active team members can view tasks for this team")
			}
		}
	}

	tasks, meta, err := s.repo.FindTasks(ctx, filter, p)
	if err != nil {
		return nil, nil, apperror.Internal("failed to fetch tasks", err)
	}
	if tasks == nil {
		tasks = []*model.Task{}
	}

	return tasks, meta, nil
}
