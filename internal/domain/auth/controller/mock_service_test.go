package controller_test

import (
	"context"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/service"
	"gdctest/pkg/pagination"

	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, ip string) error {
	args := m.Called(ctx, email, ip)
	return args.Error(0)
}

func (m *MockAuthService) SetPassword(ctx context.Context, tok, username, password, ip string) (*model.User, error) {
	args := m.Called(ctx, tok, username, password, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) ForgotPassword(ctx context.Context, email, ip string) error {
	args := m.Called(ctx, email, ip)
	return args.Error(0)
}

func (m *MockAuthService) ResetPassword(ctx context.Context, tok, password, ip string) error {
	args := m.Called(ctx, tok, password, ip)
	return args.Error(0)
}

func (m *MockAuthService) Login(ctx context.Context, email, password, ip string) (string, *model.User, error) {
	args := m.Called(ctx, email, password, ip)
	if args.Get(1) == nil {
		return args.String(0), nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*model.User), args.Error(2)
}

func (m *MockAuthService) Logout(ctx context.Context, tokenString string) error {
	args := m.Called(ctx, tokenString)
	return args.Error(0)
}

func (m *MockAuthService) ValidateJWT(ctx context.Context, tokenString, currentIP string) (*service.CustomClaims, error) {
	args := m.Called(ctx, tokenString, currentIP)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CustomClaims), args.Error(1)
}

func (m *MockAuthService) ListUsers(ctx context.Context, p pagination.Request) ([]*model.User, *pagination.Meta, error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var meta *pagination.Meta
	if args.Get(1) != nil {
		meta = args.Get(1).(*pagination.Meta)
	}
	return args.Get(0).([]*model.User), meta, args.Error(2)
}

