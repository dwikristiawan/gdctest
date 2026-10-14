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

func TestAuthController_ResetPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controller.NewAuthController(mockSvc, nil)

	t.Run("success valid reset password", func(t *testing.T) {
		req := controller.ResetPasswordRequest{
			Token:    "valid-reset-token",
			Password: "newsecurepassword",
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/reset-password", body)

		mockSvc.On("ResetPassword", mock.Anything, "valid-reset-token", "newsecurepassword", mock.Anything).
			Return(nil).Once()

		err := ctrl.ResetPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "password reset successfully")
	})

	t.Run("fails short password length", func(t *testing.T) {
		req := controller.ResetPasswordRequest{
			Token:    "valid-reset-token",
			Password: "123", // min 6
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/reset-password", body)

		err := ctrl.ResetPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("fails when token expired or invalidated", func(t *testing.T) {
		req := controller.ResetPasswordRequest{
			Token:    "expired-token",
			Password: "newsecurepassword",
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/reset-password", body)

		mockSvc.On("ResetPassword", mock.Anything, "expired-token", "newsecurepassword", mock.Anything).
			Return(apperror.New(http.StatusBadRequest, "link expired or invalidated")).Once()

		err := ctrl.ResetPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "link expired or invalidated")
	})
}
