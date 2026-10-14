package service

import (
	"context"
	"net/http"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/apperror"
	"gdctest/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) SetPassword(ctx context.Context, tok, username, password, ip string) (*model.User, error) {
	// 1. Extract email from token
	email, err := token.ExtractEmailFromToken(tok)
	if err != nil {
		return nil, apperror.New(http.StatusBadRequest, "invalid token format")
	}

	// 2. Fetch session from Redis
	session, err := s.repo.GetRegistrationSession(ctx, email)
	if err != nil {
		return nil, apperror.New(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	if session == nil || session.Token != tok {
		return nil, apperror.New(http.StatusBadRequest, "link expired or invalidated by a newer request")
	}

	// 3. Real IP Matching Check
	if session.IP != ip {
		return nil, apperror.New(http.StatusForbidden, "IP mismatch: you must set password from the same IP that requested registration")
	}

	// 4. Cek username availability
	existingUsername, err := s.repo.FindOne(ctx, model.User{Username: username})
	if err != nil {
		return nil, apperror.New(http.StatusInternalServerError, "failed to check username: "+err.Error())
	}
	if existingUsername != nil {
		return nil, apperror.New(http.StatusBadRequest, "username already taken")
	}

	// 5. Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.New(http.StatusInternalServerError, "failed to hash password: "+err.Error())
	}

	// 6. Save User into PostgreSQL
	user, err := s.repo.CreateUser(ctx, email, username, string(hashed))
	if err != nil {
		return nil, apperror.New(http.StatusInternalServerError, "failed to create user: "+err.Error())
	}

	// 7. Cleanup Redis session
	_ = s.repo.DeleteRegistrationSession(ctx, email)

	return user, nil
}
