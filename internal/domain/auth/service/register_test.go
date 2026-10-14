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

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{
		BaseURL:   "http://localhost:8080",
		JWTSecret: "secret-12345",
	})

	ctx := context.Background()
	email := "test@example.com"
	clientIP := "10.10.10.1"

	t.Run("success register zero-storage saves to redis only", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(nil, nil).Once()
		mockRepo.On("SaveRegistrationSession", ctx, email, mock.MatchedBy(func(s *model.RegistrationSession) bool {
			return s.IP == clientIP && s.Token != ""
		}), 30*time.Minute).Return(nil).Once()

		err := svc.Register(ctx, email, clientIP)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fails when email already exists", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(&model.User{ID: "existing-id", Email: email}, nil).Once()

		err := svc.Register(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("fails when rate limit exceeded 429", func(t *testing.T) {
		rdb, rmock := redismock.NewClientMock()
		limiter := ratelimit.NewLimiter(rdb)
		svcWithLimiter := service.NewAuthService(mockRepo, nil, limiter, service.AuthServiceConfig{
			BaseURL:   "http://localhost:8080",
			JWTSecret: "secret-12345",
		})

		rmock.ExpectTTL("ratelimit:ban:" + email).SetVal(25 * time.Minute)

		err := svcWithLimiter.Register(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many requests")
	})

	t.Run("fails when database query returns error 500", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(nil, errors.New("db connection failure")).Once()

		err := svc.Register(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query user")
	})

	t.Run("fails when redis save returns error 500", func(t *testing.T) {
		mockRepo.On("FindOne", ctx, model.User{Email: email}).Return(nil, nil).Once()
		mockRepo.On("SaveRegistrationSession", ctx, email, mock.Anything, 30*time.Minute).
			Return(errors.New("redis connection refused")).Once()

		err := svc.Register(ctx, email, clientIP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save registration session")
	})
}
