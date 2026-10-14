package controller_test

import (
	"net/http"
	"testing"

	"gdctest/internal/domain/auth/controller"
	"gdctest/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthController_Logout(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controller.NewAuthController(mockSvc, nil)

	t.Run("fails missing authorization header", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodPost, "/logout", nil)

		err := ctrl.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "missing or invalid authorization header")
	})

	t.Run("success valid token logout", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodPost, "/logout", nil)
		c.Request().Header.Set("Authorization", "Bearer valid-token-string")

		mockSvc.On("Logout", mock.Anything, "valid-token-string").Return(nil).Once()

		err := ctrl.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "logged out successfully")
	})

	t.Run("fails when service returns error", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodPost, "/logout", nil)
		c.Request().Header.Set("Authorization", "Bearer invalid-token-string")

		mockSvc.On("Logout", mock.Anything, "invalid-token-string").
			Return(apperror.New(http.StatusUnauthorized, "invalid or expired token")).Once()

		err := ctrl.Logout(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
