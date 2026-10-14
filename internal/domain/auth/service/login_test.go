package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/service"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{
		BaseURL:     "http://localhost:8080",
		JWTSecret:   "secret-12345",
		TokenExpiry: 1 * time.Hour,
	})

	ctx := context.Background()
	email := "loginuser@example.com"
	rawPassword := "mypassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	loginIP := "172.16.0.5"

	dbUser := &model.User{
		ID:           "user-login-uuid",
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	t.Run("fails when database query returns error 500", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(nil, errors.New("db error")).Once()

		_, _, err := svc.Login(ctx, email, rawPassword, loginIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("fails when user not found", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: "unknown@example.com"}).Return(nil, nil).Once()

		_, _, err := svc.Login(ctx, "unknown@example.com", rawPassword, loginIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email or password")
	})

	t.Run("fails when password wrong", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(dbUser, nil).Once()

		_, _, err := svc.Login(ctx, email, "wrongpass", loginIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email or password")
	})

	t.Run("succeeds and issues token", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(dbUser, nil).Once()

		tokenStr, user, err := svc.Login(ctx, email, rawPassword, loginIP)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenStr)
		assert.Equal(t, dbUser.ID, user.ID)
	})
}
