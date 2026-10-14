package service

import (
	"context"
	"database/sql"
	"strings"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) UpdateTask(ctx context.Context, currentUserID string, req UpdateTaskRequest) (*model.Task, error) {
	if req.ID == "" {
		return nil, apperror.BadRequest("task id is required")
	}

	task, err := s.repo.FindTaskByID(ctx, req.ID)
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

	// Caller must be active team member or team owner
	if team.OwnerID != currentUserID {
		member, err := s.repo.FindTeamMember(ctx, task.TeamID, currentUserID)
		if err != nil {
			return nil, apperror.Internal("failed to verify membership", err)
		}
		if member == nil || member.DeletedAt != nil {
			return nil, apperror.Forbidden("only active team members can update tasks")
		}
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return nil, apperror.BadRequest("task title cannot be empty")
		}
		task.Title = trimmed
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		if *req.Status < model.StatusBacklog || *req.Status > model.StatusArchived {
			return nil, apperror.BadRequest("invalid task status")
		}
		task.Status = *req.Status
	}

	err = s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		return s.repo.UpdateTask(ctx, tx, task)
	})
	if err != nil {
		return nil, apperror.Internal("failed to update task", err)
	}

	return task, nil
}
