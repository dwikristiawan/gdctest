package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"gdctest/internal/domain/auth/controller"
	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthController_Login(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controller.NewAuthController(mockSvc, nil)

	t.Run("success login returns token and user", func(t *testing.T) {
		req := controller.LoginRequest{
			Email:    "login@example.com",
			Password: "mypassword123",
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/login", body)

		mockSvc.On("Login", mock.Anything, "login@example.com", "mypassword123", mock.Anything).
			Return("mocked.jwt.token", &model.User{ID: "usr-456", Email: "login@example.com"}, nil).Once()

		err := ctrl.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "mocked.jwt.token")
		assert.Contains(t, rec.Body.String(), "usr-456")
	})

	t.Run("fails invalid credentials 401", func(t *testing.T) {
		req := controller.LoginRequest{
			Email:    "login@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/login", body)

		mockSvc.On("Login", mock.Anything, "login@example.com", "wrongpassword", mock.Anything).
			Return("", nil, apperror.New(http.StatusUnauthorized, "invalid email or password")).Once()

		err := ctrl.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid email or password")
	})

	t.Run("fails internal server error 500 masks internal details", func(t *testing.T) {
		req := controller.LoginRequest{
			Email:    "login@example.com",
			Password: "mypassword123",
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/login", body)

		mockSvc.On("Login", mock.Anything, "login@example.com", "mypassword123", mock.Anything).
			Return("", nil, apperror.New(http.StatusInternalServerError, "database error: connection timeout")).Once()

		err := ctrl.Login(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal Server Error. General error")
		assert.NotContains(t, rec.Body.String(), "database error: connection timeout")
	})
}

