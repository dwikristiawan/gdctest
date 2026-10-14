package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/apperror"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) Login(ctx context.Context, email, password, ip string) (string, *model.User, error) {
	user, err := s.repo.FindOne(ctx, model.User{Email: email})
	if err != nil {
		return "", nil, apperror.New(http.StatusInternalServerError, "database error: "+err.Error())
	}
	if user == nil {
		return "", nil, apperror.New(http.StatusUnauthorized, "invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, apperror.New(http.StatusUnauthorized, "invalid email or password")
	}

	// Generate JTI
	jtiBytes := make([]byte, 16)
	_, _ = rand.Read(jtiBytes)
	jti := hex.EncodeToString(jtiBytes)

	// Claims membawa Real IP
	claims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		IP: ip,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := tok.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", nil, apperror.New(http.StatusInternalServerError, "failed to sign token: "+err.Error())
	}

	return signedToken, user, nil
}
