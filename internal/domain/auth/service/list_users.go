package service

import (
	"context"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/apperror"
	"gdctest/pkg/pagination"
)

// ListUsers retrieves users using pagination.
func (s *AuthService) ListUsers(ctx context.Context, p pagination.Request) ([]*model.User, *pagination.Meta, error) {
	users, meta, err := s.repo.FindUsers(ctx, model.User{}, p)
	if err != nil {
		return nil, nil, apperror.Internal("failed to list users", err)
	}

	return users, meta, nil
}
