package service

import (
	"context"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) GetTaskDetail(ctx context.Context, taskID, currentUserID string) (*model.Task, error) {
	task, err := s.repo.FindTaskByID(ctx, taskID)
	if err != nil {
		return nil, apperror.Internal("failed to fetch task", err)
	}
	if task == nil {
		return nil, apperror.NotFound("task not found")
	}

	team, err := s.repo.FindTeamByID(ctx, task.TeamID)
	if err != nil {
		return nil, apperror.Internal("failed to fetch team", err)
	}
	if team == nil {
		return nil, apperror.NotFound("associated team not found")
	}

	// Caller must be team owner or active team member
	if team.OwnerID != currentUserID {
		member, err := s.repo.FindTeamMember(ctx, task.TeamID, currentUserID)
		if err != nil {
			return nil, apperror.Internal("failed to verify membership", err)
		}
		if member == nil || member.DeletedAt != nil {
			return nil, apperror.Forbidden("only active team members can view task details")
		}
	}

	collaborators, err := s.repo.FindTaskCollaborators(ctx, taskID)
	if err != nil {
		return nil, apperror.Internal("failed to fetch task collaborators", err)
	}
	if collaborators == nil {
		collaborators = []*model.TaskCollaborator{}
	}
	task.Collaborators = collaborators

	return task, nil
}
