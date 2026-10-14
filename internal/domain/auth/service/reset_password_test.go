package service_test

import (
	"context"
	"errors"
	"testing"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/service"
	"gdctest/pkg/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_ResetPassword(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{
		BaseURL:   "http://localhost:8080",
		JWTSecret: "secret-12345",
	})

	ctx := context.Background()
	email := "reset@example.com"
	tok, _ := token.GenerateRegistrationToken(email)
	ip := "127.0.0.1"

	session := &model.RegistrationSession{Token: tok, IP: ip}

	t.Run("fails when token format invalid", func(t *testing.T) {
		err := svc.ResetPassword(ctx, "invalid-token", "newpass123", ip)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("fails when session expired or invalidated", func(t *testing.T) {
		mockRepo.On("GetResetPasswordSession", ctx, email).Return(nil, nil).Once()

		err := svc.ResetPassword(ctx, tok, "newpass123", ip)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "link expired or invalidated")
	})

	t.Run("fails when IP mismatch", func(t *testing.T) {
		mockRepo.On("GetResetPasswordSession", ctx, email).Return(session, nil).Once()

		err := svc.ResetPassword(ctx, tok, "newpass123", "10.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "IP mismatch")
	})

	t.Run("fails when database update password returns error 500", func(t *testing.T) {
		mockRepo.On("GetResetPasswordSession", ctx, email).Return(session, nil).Once()
		mockRepo.On("UpdateUser", ctx, model.User{Email: email}, mock.Anything).
			Return(errors.New("db error on update password")).Once()

		err := svc.ResetPassword(ctx, tok, "newpass123", ip)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update password")
	})

	t.Run("succeeds and clears all cache", func(t *testing.T) {
		mockRepo.On("GetResetPasswordSession", ctx, email).Return(session, nil).Once()
		mockRepo.On("UpdateUser", ctx, model.User{Email: email}, mock.Anything).Return(nil).Once()
		mockRepo.On("ClearAllUserSessions", ctx, email).Return(nil).Once()

		err := svc.ResetPassword(ctx, tok, "newpass123", ip)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
