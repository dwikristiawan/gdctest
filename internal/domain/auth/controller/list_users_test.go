package controller_test

import (
	"errors"
	"net/http"
	"testing"

	"gdctest/internal/domain/auth/controller"
	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/apperror"
	"gdctest/pkg/pagination"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthController_ListUsers(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controller.NewAuthController(mockSvc, nil)

	t.Run("success returns items and meta", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodGet, "/v1.0/auth/users?page=1&limit=10&order=created_at:desc", nil)

		expectedUsers := []*model.User{
			{ID: "usr-1", Email: "u1@example.com", Username: "user1"},
			{ID: "usr-2", Email: "u2@example.com", Username: "user2"},
		}
		expectedMeta := &pagination.Meta{
			Page:       1,
			Limit:      10,
			TotalRows:  2,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		}

		mockSvc.On("ListUsers", mock.Anything, pagination.Request{
			Page:  1,
			Limit: 10,
			Order: "created_at:desc",
		}).Return(expectedUsers, expectedMeta, nil).Once()

		err := ctrl.ListUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "u1@example.com")
		assert.Contains(t, rec.Body.String(), "u2@example.com")
		assert.Contains(t, rec.Body.String(), `"total_rows":2`)
	})

	t.Run("service internal error returns 500 masked message", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodGet, "/v1.0/auth/users", nil)

		mockSvc.On("ListUsers", mock.Anything, mock.Anything).
			Return(nil, nil, apperror.Internal("failed to list users", errors.New("db connection failure"))).Once()

		err := ctrl.ListUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal Server Error. General error")
	})
}
