package service

import (
	"context"
	"net/http"
	"time"

	"gdctest/pkg/apperror"
)

func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.ValidateJWT(ctx, tokenString, "")
	if err != nil {
		return err
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		if err := s.repo.BlacklistToken(ctx, claims.ID, remaining); err != nil {
			return apperror.New(http.StatusInternalServerError, "failed to blacklist token: "+err.Error())
		}
	}

	return nil
}
