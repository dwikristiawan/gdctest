package service

import (
	"context"
	"database/sql"

	"gdctest/pkg/apperror"
)

func (s *TaskService) DeleteTeam(ctx context.Context, teamID, userID string) error {
	team, err := s.repo.FindTeamByID(ctx, teamID)
	if err != nil {
		return apperror.Internal("failed to fetch team", err)
	}
	if team == nil {
		return apperror.NotFound("team not found")
	}

	if team.OwnerID != userID {
		return apperror.Forbidden("only team owner can delete the team")
	}

	err = s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		return s.repo.DeleteTeam(ctx, tx, teamID)
	})
	if err != nil {
		return apperror.Internal("failed to delete team", err)
	}

	return nil
}
