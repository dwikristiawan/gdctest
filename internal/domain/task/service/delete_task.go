package service

import (
	"context"
	"database/sql"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) DeleteTask(ctx context.Context, taskID, currentUserID string) error {
	task, err := s.repo.FindTaskByID(ctx, taskID)
	if err != nil {
		return apperror.Internal("failed to fetch task", err)
	}
	if task == nil {
		return apperror.NotFound("task not found")
	}

	team, err := s.repo.FindTeamByID(ctx, task.TeamID)
	if err != nil {
		return apperror.Internal("failed to fetch team", err)
	}
	if team == nil {
		return apperror.NotFound("associated team not found")
	}

	// Permission: pembuat task, owner team, atau admin team
	isTaskCreator := task.OwnerID == currentUserID
	isTeamOwner := team.OwnerID == currentUserID

	if !isTaskCreator && !isTeamOwner {
		member, err := s.repo.FindTeamMember(ctx, task.TeamID, currentUserID)
		if err != nil {
			return apperror.Internal("failed to verify membership", err)
		}
		if member == nil || member.DeletedAt != nil || member.RoleType != model.RoleTeamAdmin {
			return apperror.Forbidden("only task creator or team admin can delete task")
		}
	}

	err = s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		return s.repo.DeleteTask(ctx, tx, taskID)
	})
	if err != nil {
		return apperror.Internal("failed to delete task", err)
	}

	return nil
}
