package service

import (
	"context"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

func (s *TaskService) GetTeamMembers(ctx context.Context, teamID, currentUserID string) ([]*model.TeamMember, error) {
	team, err := s.repo.FindTeamByID(ctx, teamID)
	if err != nil {
		return nil, apperror.Internal("failed to fetch team", err)
	}
	if team == nil {
		return nil, apperror.NotFound("team not found")
	}

	// Caller must be owner or active team member
	if team.OwnerID != currentUserID {
		callerMember, err := s.repo.FindTeamMember(ctx, teamID, currentUserID)
		if err != nil {
			return nil, apperror.Internal("failed to verify membership", err)
		}
		if callerMember == nil || callerMember.DeletedAt != nil {
			return nil, apperror.Forbidden("only team members can view the member list")
		}
	}

	members, err := s.repo.FindActiveTeamMembers(ctx, teamID)
	if err != nil {
		return nil, apperror.Internal("failed to fetch team members", err)
	}
	if members == nil {
		members = []*model.TeamMember{}
	}
	return members, nil
}
