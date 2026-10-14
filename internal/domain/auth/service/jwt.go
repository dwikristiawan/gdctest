package service

import (
	"context"
	"errors"
	"net/http"

	"gdctest/pkg/apperror"

	"github.com/golang-jwt/jwt/v5"
)

func (s *AuthService) ValidateJWT(ctx context.Context, tokenString, currentIP string) (*CustomClaims, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !tok.Valid {
		return nil, apperror.New(http.StatusUnauthorized, "invalid or expired token")
	}

	claims, ok := tok.Claims.(*CustomClaims)
	if !ok {
		return nil, apperror.New(http.StatusUnauthorized, "invalid token claims")
	}

	// Validasi Real IP jika currentIP diberikan
	if currentIP != "" && claims.IP != "" && claims.IP != currentIP {
		return nil, apperror.New(http.StatusForbidden, "IP mismatch: request IP does not match login IP")
	}

	// Cek blacklist di Redis
	isBlacklisted, err := s.repo.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, apperror.New(http.StatusInternalServerError, "failed to verify token revocation")
	}
	if isBlacklisted {
		return nil, apperror.New(http.StatusUnauthorized, "token has been revoked")
	}

	return claims, nil
}
