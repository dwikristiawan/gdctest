package service_test

import (
	"context"
	"errors"
	"testing"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/service"
	"gdctest/pkg/apperror"
	"gdctest/pkg/pagination"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_ListUsers(t *testing.T) {
	ctx := context.Background()
	req := pagination.Request{
		Page:  1,
		Limit: 10,
	}

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockAuthRepository)
		svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{})

		expectedUsers := []*model.User{
			{ID: "uuid-1", Email: "user1@example.com", Username: "user1"},
			{ID: "uuid-2", Email: "user2@example.com", Username: "user2"},
		}
		expectedMeta := &pagination.Meta{
			Page:       1,
			Limit:      10,
			TotalRows:  2,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		}

		mockRepo.On("FindUsers", ctx, model.User{}, req).Return(expectedUsers, expectedMeta, nil)

		users, meta, err := svc.ListUsers(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		assert.Equal(t, expectedMeta, meta)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repo error returns internal error", func(t *testing.T) {
		mockRepo := new(MockAuthRepository)
		svc := service.NewAuthService(mockRepo, nil, nil, service.AuthServiceConfig{})

		mockRepo.On("FindUsers", ctx, model.User{}, req).Return(nil, nil, errors.New("db error"))

		users, meta, err := svc.ListUsers(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Nil(t, meta)

		var appErr *apperror.Error
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, 500, appErr.Code)
		mockRepo.AssertExpectations(t)
	})
}
