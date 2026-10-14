package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_ValidateJWT(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{
		BaseURL:     "http://localhost:8080",
		JWTSecret:   "secret-12345",
		TokenExpiry: 1 * time.Hour,
	})

	ctx := context.Background()
	email := "jwtuser@example.com"
	rawPassword := "mypassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	loginIP := "192.168.1.100"

	dbUser := &model.User{
		ID:           "jwt-user-uuid",
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(dbUser, nil).Once()
	tokenStr, _, err := svc.Login(ctx, email, rawPassword, loginIP)
	assert.NoError(t, err)

	t.Run("fails when token string corrupted or invalid", func(t *testing.T) {
		_, err := svc.ValidateJWT(ctx, "corrupted.token.payload", loginIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired token")
	})

	t.Run("fails when IP mismatch", func(t *testing.T) {
		_, err := svc.ValidateJWT(ctx, tokenStr, "192.168.1.200")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "IP mismatch")
	})

	t.Run("fails when redis error during blacklist verification 500", func(t *testing.T) {
		mockRepo.On("IsTokenBlacklisted", ctx, mock.Anything).Return(false, errors.New("redis down")).Once()

		_, err := svc.ValidateJWT(ctx, tokenStr, loginIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify token revocation")
	})

	t.Run("fails when token blacklisted", func(t *testing.T) {
		mockRepo.On("IsTokenBlacklisted", ctx, mock.Anything).Return(true, nil).Once()

		_, err := svc.ValidateJWT(ctx, tokenStr, loginIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "revoked")
	})

	t.Run("succeeds with matching IP and valid token", func(t *testing.T) {
		mockRepo.On("IsTokenBlacklisted", ctx, mock.Anything).Return(false, nil).Once()

		claims, err := svc.ValidateJWT(ctx, tokenStr, loginIP)
		assert.NoError(t, err)
		assert.Equal(t, dbUser.ID, claims.Subject)
		assert.Equal(t, loginIP, claims.IP)
	})
}
