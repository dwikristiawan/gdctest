package service

import (
	"context"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/repository"
	"gdctest/pkg/infra"
	"gdctest/pkg/pagination"
	"gdctest/pkg/ratelimit"

	"github.com/golang-jwt/jwt/v5"
)

type AuthServiceConfig struct {
	BaseURL     string
	JWTSecret   string
	TokenExpiry time.Duration
}

type CustomClaims struct {
	jwt.RegisteredClaims
	IP string `json:"ip"`
}

type AuthServiceInterface interface {
	Register(ctx context.Context, email, ip string) error
	SetPassword(ctx context.Context, tok, username, password, ip string) (*model.User, error)
	ForgotPassword(ctx context.Context, email, ip string) error
	ResetPassword(ctx context.Context, tok, password, ip string) error
	Login(ctx context.Context, email, password, ip string) (string, *model.User, error)
	Logout(ctx context.Context, tokenString string) error
	ValidateJWT(ctx context.Context, tokenString, currentIP string) (*CustomClaims, error)
	ListUsers(ctx context.Context, p pagination.Request) ([]*model.User, *pagination.Meta, error)
}

type AuthService struct {
	repo    repository.AuthRepositoryInterface
	mailer  *infra.Mailer
	limiter *ratelimit.Limiter
	cfg     AuthServiceConfig
}

func NewAuthService(
	repo repository.AuthRepositoryInterface,
	mailer *infra.Mailer,
	limiter *ratelimit.Limiter,
	cfg AuthServiceConfig,
) AuthServiceInterface {
	if cfg.TokenExpiry == 0 {
		cfg.TokenExpiry = 24 * time.Hour
	}
	return &AuthService{
		repo:    repo,
		mailer:  mailer,
		limiter: limiter,
		cfg:     cfg,
	}
}
