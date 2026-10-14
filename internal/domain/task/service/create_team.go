package service

import (
	"context"
	"database/sql"
	"strings"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) CreateTeam(ctx context.Context, ownerID string, req CreateTeamRequest) (*model.Team, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperror.BadRequest("team name is required")
	}

	team := &model.Team{
		Name:        name,
		Description: req.Description,
		OwnerID:     ownerID,
	}

	var createdTeam *model.Team
	err := s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		var txErr error
		createdTeam, txErr = s.repo.CreateTeam(ctx, tx, team)
		if txErr != nil {
			return txErr
		}

		// Otomatis daftarkan owner sebagai member admin
		ownerMember := &model.TeamMember{
			TeamID:   createdTeam.ID,
			UserID:   ownerID,
			RoleType: model.RoleTeamAdmin,
		}
		return s.repo.AddTeamMember(ctx, tx, ownerMember)
	})
	if err != nil {
		return nil, apperror.Internal("failed to create team", err)
	}

	return createdTeam, nil
}
