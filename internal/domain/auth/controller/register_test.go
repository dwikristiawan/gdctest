package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"gdctest/internal/domain/auth/controller"
	"gdctest/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthController_Register(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controller.NewAuthController(mockSvc, nil)

	t.Run("success valid registration", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "valid@example.com"})
		c, rec := setupEchoTest(http.MethodPost, "/register", body)

		mockSvc.On("Register", mock.Anything, "valid@example.com", mock.Anything).Return(nil).Once()

		err := ctrl.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "verification email sent")
	})

	t.Run("fails invalid email format", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "not-an-email"})
		c, rec := setupEchoTest(http.MethodPost, "/register", body)

		err := ctrl.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("fails when service returns error", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "taken@example.com"})
		c, rec := setupEchoTest(http.MethodPost, "/register", body)

		mockSvc.On("Register", mock.Anything, "taken@example.com", mock.Anything).
			Return(apperror.New(http.StatusBadRequest, "email is already registered")).Once()

		err := ctrl.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "email is already registered")
	})

	t.Run("fails when service returns internal server error 500", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "dberr@example.com"})
		c, rec := setupEchoTest(http.MethodPost, "/register", body)

		mockSvc.On("Register", mock.Anything, "dberr@example.com", mock.Anything).
			Return(apperror.New(http.StatusInternalServerError, "database connection error")).Once()

		err := ctrl.Register(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		// Masked general error returned to client
		assert.Contains(t, rec.Body.String(), "Internal Server Error. General error")
		assert.NotContains(t, rec.Body.String(), "database connection error")
	})
}

