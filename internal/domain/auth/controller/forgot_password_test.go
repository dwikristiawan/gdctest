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

func TestAuthController_ForgotPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controller.NewAuthController(mockSvc, nil)

	t.Run("success valid forgot password request", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "user@example.com"})
		c, rec := setupEchoTest(http.MethodPost, "/forgot-password", body)

		mockSvc.On("ForgotPassword", mock.Anything, "user@example.com", mock.Anything).Return(nil).Once()

		err := ctrl.ForgotPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "reset password email sent")
	})

	t.Run("fails invalid email syntax", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "invalid"})
		c, rec := setupEchoTest(http.MethodPost, "/forgot-password", body)

		err := ctrl.ForgotPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("fails when service returns not found", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "notfound@example.com"})
		c, rec := setupEchoTest(http.MethodPost, "/forgot-password", body)

		mockSvc.On("ForgotPassword", mock.Anything, "notfound@example.com", mock.Anything).
			Return(apperror.New(http.StatusBadRequest, "email not found")).Once()

		err := ctrl.ForgotPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "email not found")
	})
}
