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

func TestAuthController_SetPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controller.NewAuthController(mockSvc, nil)

	t.Run("success valid set password", func(t *testing.T) {
		req := controller.SetPasswordRequest{
			Token:    "sample-token",
			Username: "validuser",
			Password: "securepassword",
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/set-password", body)

		mockSvc.On("SetPassword", mock.Anything, "sample-token", "validuser", "securepassword", mock.Anything).
			Return(&model.User{ID: "usr-123", Username: "validuser"}, nil).Once()

		err := ctrl.SetPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "usr-123")
	})

	t.Run("fails invalid validation rules", func(t *testing.T) {
		req := controller.SetPasswordRequest{
			Token:    "", // required
			Username: "a", // min 3
			Password: "123", // min 6
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/set-password", body)

		err := ctrl.SetPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("fails when service returns IP mismatch 403", func(t *testing.T) {
		req := controller.SetPasswordRequest{
			Token:    "sample-token",
			Username: "validuser",
			Password: "securepassword",
		}
		body, _ := json.Marshal(req)
		c, rec := setupEchoTest(http.MethodPost, "/set-password", body)

		mockSvc.On("SetPassword", mock.Anything, "sample-token", "validuser", "securepassword", mock.Anything).
			Return(nil, apperror.New(http.StatusForbidden, "IP mismatch")).Once()

		err := ctrl.SetPassword(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "IP mismatch")
	})
}
