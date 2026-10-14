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

func TestAuthService_SetPassword(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{
		BaseURL:   "http://localhost:8080",
		JWTSecret: "secret-12345",
	})

	ctx := context.Background()
	email := "valid@example.com"
	tok, _ := token.GenerateRegistrationToken(email)
	correctIP := "192.168.1.10"
	wrongIP := "203.0.113.195"

	session := &model.RegistrationSession{Token: tok, IP: correctIP}

	t.Run("fails when token format invalid", func(t *testing.T) {
		_, err := svc.SetPassword(ctx, "invalid-token-string", "username1", "pass123", correctIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("fails when session expired or invalidated", func(t *testing.T) {
		mockRepo.On("GetRegistrationSession", ctx, email).Return(nil, nil).Once()

		_, err := svc.SetPassword(ctx, tok, "username1", "pass123", correctIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "link expired or invalidated")
	})

	t.Run("fails when IP mismatch", func(t *testing.T) {
		mockRepo.On("GetRegistrationSession", ctx, email).Return(session, nil).Once()

		_, err := svc.SetPassword(ctx, tok, "username1", "pass123", wrongIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "IP mismatch")
	})

	t.Run("fails when username already taken", func(t *testing.T) {
		mockRepo.On("GetRegistrationSession", ctx, email).Return(session, nil).Once()
		mockRepo.On("FindOne", ctx, model.User{Username: "taken_user"}).
			Return(&model.User{ID: "existing-user-id", Username: "taken_user"}, nil).Once()

		_, err := svc.SetPassword(ctx, tok, "taken_user", "pass123", correctIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username already taken")
	})

	t.Run("fails when database create user fails 500", func(t *testing.T) {
		mockRepo.On("GetRegistrationSession", ctx, email).Return(session, nil).Once()
		mockRepo.On("FindOne", ctx, model.User{Username: "username1"}).Return(nil, nil).Once()
		mockRepo.On("CreateUser", ctx, email, "username1", mock.Anything).
			Return(nil, errors.New("db insert deadlock")).Once()

		_, err := svc.SetPassword(ctx, tok, "username1", "pass123", correctIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
	})

	t.Run("succeeds when IP matches and persists user", func(t *testing.T) {
		mockRepo.On("GetRegistrationSession", ctx, email).Return(session, nil).Once()
		mockRepo.On("FindOne", ctx, model.User{Username: "username1"}).Return(nil, nil).Once()
		mockRepo.On("CreateUser", ctx, email, "username1", mock.Anything).Return(&model.User{
			ID:       "uuid-created",
			Username: "username1",
		}, nil).Once()
		mockRepo.On("DeleteRegistrationSession", ctx, email).Return(nil).Once()

		user, err := svc.SetPassword(ctx, tok, "username1", "pass123", correctIP)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "uuid-created", user.ID)
	})
}
