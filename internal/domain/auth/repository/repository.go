package repository

import (
	"context"
	"database/sql"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/pagination"

	"github.com/redis/go-redis/v9"
)

type AuthRepositoryInterface interface {
	// PostgreSQL
	FindOne(ctx context.Context, filter model.User) (*model.User, error)
	FindUsers(ctx context.Context, filter model.User, p pagination.Request) ([]*model.User, *pagination.Meta, error)
	CreateUser(ctx context.Context, email, username, passwordHash string) (*model.User, error)
	UpdateUser(ctx context.Context, where model.User, update model.User) error

	// Redis Registration Session
	SaveRegistrationSession(ctx context.Context, email string, session *model.RegistrationSession, ttl time.Duration) error
	GetRegistrationSession(ctx context.Context, email string) (*model.RegistrationSession, error)
	DeleteRegistrationSession(ctx context.Context, email string) error

	// Redis Reset Password Session
	SaveResetPasswordSession(ctx context.Context, email string, session *model.RegistrationSession, ttl time.Duration) error
	GetResetPasswordSession(ctx context.Context, email string) (*model.RegistrationSession, error)
	DeleteResetPasswordSession(ctx context.Context, email string) error
	ClearAllUserSessions(ctx context.Context, email string) error

	// Redis JWT Blacklist
	BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

// AuthRepository embeds both PostgresUserRepository and RedisAuthRepository
type AuthRepository struct {
	*PostgresUserRepository
	*RedisAuthRepository
}

func NewAuthRepository(db *sql.DB, rdb *redis.Client) AuthRepositoryInterface {
	return &AuthRepository{
		PostgresUserRepository: NewPostgresUserRepository(db),
		RedisAuthRepository:    NewRedisAuthRepository(rdb),
	}
}
