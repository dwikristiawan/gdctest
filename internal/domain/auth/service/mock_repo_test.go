package service_test

import (
	"context"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/pagination"

	"github.com/stretchr/testify/mock"
)

// MockAuthRepository is a Mockito-style mock struct using testify/mock
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) FindOne(ctx context.Context, filter model.User) (*model.User, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) FindUsers(ctx context.Context, filter model.User, p pagination.Request) ([]*model.User, *pagination.Meta, error) {
	args := m.Called(ctx, filter, p)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var meta *pagination.Meta
	if args.Get(1) != nil {
		meta = args.Get(1).(*pagination.Meta)
	}
	return args.Get(0).([]*model.User), meta, args.Error(2)
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, email, username, passwordHash string) (*model.User, error) {
	args := m.Called(ctx, email, username, passwordHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthRepository) UpdateUser(ctx context.Context, where model.User, update model.User) error {
	args := m.Called(ctx, where, update)
	return args.Error(0)
}

func (m *MockAuthRepository) SaveRegistrationSession(ctx context.Context, email string, session *model.RegistrationSession, ttl time.Duration) error {
	args := m.Called(ctx, email, session, ttl)
	return args.Error(0)
}

func (m *MockAuthRepository) GetRegistrationSession(ctx context.Context, email string) (*model.RegistrationSession, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RegistrationSession), args.Error(1)
}

func (m *MockAuthRepository) DeleteRegistrationSession(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthRepository) SaveResetPasswordSession(ctx context.Context, email string, session *model.RegistrationSession, ttl time.Duration) error {
	args := m.Called(ctx, email, session, ttl)
	return args.Error(0)
}

func (m *MockAuthRepository) GetResetPasswordSession(ctx context.Context, email string) (*model.RegistrationSession, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RegistrationSession), args.Error(1)
}

func (m *MockAuthRepository) DeleteResetPasswordSession(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthRepository) ClearAllUserSessions(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthRepository) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	args := m.Called(ctx, jti, ttl)
	return args.Error(0)
}

func (m *MockAuthRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}
