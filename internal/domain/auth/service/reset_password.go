package service

import (
	"context"
	"net/http"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/apperror"
	"gdctest/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) ResetPassword(ctx context.Context, tok, password, ip string) error {
	// 1. Extract email from token
	email, err := token.ExtractEmailFromToken(tok)
	if err != nil {
		return apperror.New(http.StatusBadRequest, "invalid token format")
	}

	// 2. Fetch session from Redis
	session, err := s.repo.GetResetPasswordSession(ctx, email)
	if err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	if session == nil || session.Token != tok {
		return apperror.New(http.StatusBadRequest, "link expired or invalidated by a newer request")
	}

	// 3. Real IP Matching Check
	if session.IP != ip {
		return apperror.New(http.StatusForbidden, "IP mismatch: you must reset password from the same IP that requested it")
	}

	// 4. Hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to hash password: "+err.Error())
	}

	// 5. Update user password di PostgreSQL via UpdateUser
	if err := s.repo.UpdateUser(ctx, model.User{Email: email}, model.User{PasswordHash: string(hashed)}); err != nil {
		return apperror.New(http.StatusInternalServerError, "failed to update password: "+err.Error())
	}

	// 6. Kalau sukses: SEMUA data email dihilangkan dari cache (reg, reset, rate limit)
	_ = s.repo.ClearAllUserSessions(ctx, email)

	return nil
}
