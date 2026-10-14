package service_test

import (
	"context"
	"testing"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Logout(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{
		BaseURL:     "http://localhost:8080",
		JWTSecret:   "secret-12345",
		TokenExpiry: 1 * time.Hour,
	})

	ctx := context.Background()
	email := "logoutuser@example.com"
	rawPassword := "mypassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

	dbUser := &model.User{
		ID:           "logout-uuid",
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(dbUser, nil).Once()
	tokenStr, _, err := svc.Login(ctx, email, rawPassword, "10.0.0.1")
	assert.NoError(t, err)

	t.Run("succeeds and blacklists token", func(t *testing.T) {
		mockRepo.On("IsTokenBlacklisted", ctx, mock.Anything).Return(false, nil).Once()
		mockRepo.On("BlacklistToken", ctx, mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.Logout(ctx, tokenStr)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
