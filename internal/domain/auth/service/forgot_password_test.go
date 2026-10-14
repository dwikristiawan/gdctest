package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/service"
	"gdctest/pkg/ratelimit"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_ForgotPassword(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{
		BaseURL:   "http://localhost:8080",
		JWTSecret: "secret-12345",
	})

	ctx := context.Background()
	email := "user@example.com"
	clientIP := "192.168.1.50"

	t.Run("fails when rate limit exceeded 429", func(t *testing.T) {
		rdb, rmock := redismock.NewClientMock()
		limiter := ratelimit.NewLimiter(rdb)
		svcWithLimiter := service.NewAuthService(mockRepo, nil, limiter, service.AuthServiceConfig{
			BaseURL:   "http://localhost:8080",
			JWTSecret: "secret-12345",
		})

		rmock.ExpectTTL("ratelimit:ban:" + email).SetVal(15 * time.Minute)

		err := svcWithLimiter.ForgotPassword(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many requests")
	})

	t.Run("fails when email not found", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(nil, nil).Once()

		err := svc.ForgotPassword(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email not found")
	})

	t.Run("fails when database query returns error 500", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(nil, errors.New("db query error")).Once()

		err := svc.ForgotPassword(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("fails when redis save reset session returns error 500", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(&model.User{ID: "usr-1", Email: email}, nil).Once()
		mockRepo.On("SaveResetPasswordSession", ctx, email, mock.Anything, 15*time.Minute).
			Return(errors.New("redis timeout")).Once()

		err := svc.ForgotPassword(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save reset password session")
	})

	t.Run("succeeds when user exists and saves 15m session", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(&model.User{ID: "usr-1", Email: email}, nil).Once()
		mockRepo.On("SaveResetPasswordSession", ctx, email, mock.MatchedBy(func(s *model.RegistrationSession) bool {
			return s.IP == clientIP && s.Token != ""
		}), 15*time.Minute).Return(nil).Once()

		err := svc.ForgotPassword(ctx, email, clientIP)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
