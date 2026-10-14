package service

import (
	"context"
	"database/sql"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) AddMember(ctx context.Context, teamID, currentUserID string, req AddMemberRequest) error {
	if req.UserID == "" {
		return apperror.BadRequest("user_id is required")
	}
	if req.RoleType != model.RoleTeamAdmin && req.RoleType != model.RoleTeamMember {
		return apperror.BadRequest("invalid role_type: must be 0 (Admin) or 1 (Member)")
	}

	team, err := s.repo.FindTeamByID(ctx, teamID)
	if err != nil {
		return apperror.Internal("failed to fetch team", err)
	}
	if team == nil {
		return apperror.NotFound("team not found")
	}

	// Cek otoritas current user: harus owner atau team admin aktif
	if team.OwnerID != currentUserID {
		callerMember, err := s.repo.FindTeamMember(ctx, teamID, currentUserID)
		if err != nil {
			return apperror.Internal("failed to verify membership", err)
		}
		if callerMember == nil || callerMember.DeletedAt != nil || callerMember.RoleType != model.RoleTeamAdmin {
			return apperror.Forbidden("only team owner or admin can add members")
		}
	}

	// Cek status keanggotaan target user saat ini
	targetMember, err := s.repo.FindTeamMember(ctx, teamID, req.UserID)
	if err != nil {
		return apperror.Internal("failed to check existing member", err)
	}
	if targetMember != nil && targetMember.DeletedAt == nil {
		return apperror.BadRequest("user is already an active member of this team")
	}

	member := &model.TeamMember{
		TeamID:   teamID,
		UserID:   req.UserID,
		RoleType: req.RoleType,
	}

	err = s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		return s.repo.AddTeamMember(ctx, tx, member)
	})
	if err != nil {
		return apperror.Internal("failed to add team member", err)
	}

	return nil
}
