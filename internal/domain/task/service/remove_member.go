package service

import (
	"context"
	"database/sql"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) RemoveMember(ctx context.Context, teamID, currentUserID, targetUserID string) error {
	if targetUserID == "" {
		return apperror.BadRequest("target user_id is required")
	}

	team, err := s.repo.FindTeamByID(ctx, teamID)
	if err != nil {
		return apperror.Internal("failed to fetch team", err)
	}
	if team == nil {
		return apperror.NotFound("team not found")
	}

	// Owner tidak boleh dihapus
	if team.OwnerID == targetUserID {
		return apperror.BadRequest("cannot remove team owner from team")
	}

	// Cek otorisasi
	isSelf := currentUserID == targetUserID
	isOwner := team.OwnerID == currentUserID

	if !isSelf && !isOwner {
		callerMember, err := s.repo.FindTeamMember(ctx, teamID, currentUserID)
		if err != nil {
			return apperror.Internal("failed to verify caller membership", err)
		}
		if callerMember == nil || callerMember.DeletedAt != nil || callerMember.RoleType != model.RoleTeamAdmin {
			return apperror.Forbidden("only team owner or admin can remove members")
		}
	}

	// Cek apakah target saat ini aktif
	targetMember, err := s.repo.FindTeamMember(ctx, teamID, targetUserID)
	if err != nil {
		return apperror.Internal("failed to find target member", err)
	}
	if targetMember == nil || targetMember.DeletedAt != nil {
		return apperror.NotFound("member not found or already removed")
	}

	err = s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		return s.repo.SoftDeleteTeamMember(ctx, tx, teamID, targetUserID)
	})
	if err != nil {
		return apperror.Internal("failed to soft delete team member", err)
	}

	return nil
}
