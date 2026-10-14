package controller

import (
	"gdctest/internal/domain/auth/service"
	"gdctest/pkg/echo"
	"gdctest/pkg/ratelimit"
)

type AuthControllerInterface interface {
	Register(c echo.Context) error
	SetPassword(c echo.Context) error
	ForgotPassword(c echo.Context) error
	ResetPassword(c echo.Context) error
	Login(c echo.Context) error
	Logout(c echo.Context) error
	ListUsers(c echo.Context) error
}

type AuthController struct {
	svc service.AuthServiceInterface
}

func NewAuthController(svc service.AuthServiceInterface, _ ...*ratelimit.Limiter) AuthControllerInterface {
	return &AuthController{svc: svc}
}
