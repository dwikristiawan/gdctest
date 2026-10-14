package middleware

import (
	"errors"
	"net/http"
	"strings"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
	"gdctest/pkg/httputil"

	"github.com/golang-jwt/jwt/v5"
	echoPkg "github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email,omitempty"`
	IP    string `json:"ip,omitempty"`
}

// Auth verifies bearer JWT token and populates user_id into echo.Context
func Auth(jwtSecret string, rdb *redis.Client) echoPkg.MiddlewareFunc {
	return func(next echoPkg.HandlerFunc) echoPkg.HandlerFunc {
		return func(c echoPkg.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.Respond(c, nil, apperror.New(http.StatusUnauthorized, "unauthorized"))
			}

			tokenString := parts[1]
			tok, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !tok.Valid {
				return echo.Respond(c, nil, apperror.New(http.StatusUnauthorized, "invalid or expired token"))
			}

			claims, ok := tok.Claims.(*JWTClaims)
			if !ok {
				return echo.Respond(c, nil, apperror.New(http.StatusUnauthorized, "invalid token claims"))
			}

			// Validate IP mismatch if claim has IP
			currentIP := httputil.GetRealIP(c)
			if currentIP != "" && claims.IP != "" && claims.IP != currentIP {
				return echo.Respond(c, nil, apperror.New(http.StatusForbidden, "IP mismatch: request IP does not match login IP"))
			}

			// Check Redis blacklist
			if rdb != nil && claims.ID != "" {
				exists, err := rdb.Exists(c.Request().Context(), "auth:blacklist:"+claims.ID).Result()
				if err != nil {
					return echo.Respond(c, nil, apperror.New(http.StatusInternalServerError, "failed to verify token revocation"))
				}
				if exists > 0 {
					return echo.Respond(c, nil, apperror.New(http.StatusUnauthorized, "token has been revoked"))
				}
			}

			c.Set("user_id", claims.Subject)
			return next(c)
		}
	}
}
