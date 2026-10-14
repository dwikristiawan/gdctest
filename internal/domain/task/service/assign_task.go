package service

import (
	"context"
	"database/sql"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
	"gdctest/pkg/logger"
)

func (s *TaskService) AssignTask(ctx context.Context, currentUserID string, req AssignTaskRequest) error {
	if req.TaskID == "" {
		return apperror.BadRequest("task_id is required")
	}
	if req.UserID == "" {
		return apperror.BadRequest("user_id is required")
	}
	if req.RoleType < model.RoleAssignee || req.RoleType > model.RoleWatcher {
		return apperror.BadRequest("invalid collaborator role_type: must be 0 (Assignee), 1 (Reviewer), or 2 (Watcher)")
	}

	task, err := s.repo.FindTaskByID(ctx, req.TaskID)
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

	// Caller must be active team member or team owner
	if team.OwnerID != currentUserID {
		callerMember, err := s.repo.FindTeamMember(ctx, task.TeamID, currentUserID)
		if err != nil {
			return apperror.Internal("failed to verify caller membership", err)
		}
		if callerMember == nil || callerMember.DeletedAt != nil {
			return apperror.Forbidden("only active team members can assign tasks")
		}
	}

	// Target user MUST be an active team member of the task's team
	targetMember, err := s.repo.FindTeamMember(ctx, task.TeamID, req.UserID)
	if err != nil {
		return apperror.Internal("failed to verify target membership", err)
	}
	if targetMember == nil || targetMember.DeletedAt != nil {
		return apperror.BadRequest("assigned user must be an active member of the team")
	}

	var prevAssigneeID *string
	if req.RoleType == model.RoleAssignee {
		existingCols, _ := s.repo.FindTaskCollaborators(ctx, req.TaskID)
		for _, ec := range existingCols {
			if ec.RoleType == model.RoleAssignee {
				val := ec.UserID
				prevAssigneeID = &val
				break
			}
		}
	}

	col := &model.TaskCollaborator{
		TaskID:    req.TaskID,
		UserID:    req.UserID,
		RoleType:  req.RoleType,
		UpdatedBy: &currentUserID,
	}

	log := &model.TaskLog{
		TaskID:             req.TaskID,
		PreviousAssigneeID: prevAssigneeID,
		NewAssigneeID:      req.UserID,
		ChangedBy:          currentUserID,
		Action:             "ASSIGN",
	}

	err = s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		if err := s.repo.AssignTaskCollaborator(ctx, tx, col); err != nil {
			return err
		}
		if err := s.repo.CreateTaskLog(ctx, tx, log); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return apperror.Internal("failed to assign collaborator", err)
	}

	logger.FromContext(ctx).Info("notification sent: task assigned",
		"task_id", req.TaskID,
		"assignee", req.UserID,
		"changed_by", currentUserID,
	)
	return nil
}
