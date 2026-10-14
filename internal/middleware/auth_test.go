package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gdctest/internal/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEchoTest(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func generateTestJWT(secret, subject, ip, jti string, exp time.Duration) string {
	claims := &middleware.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
		},
		IP: ip,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := tok.SignedString([]byte(secret))
	return tokenStr
}

func TestAuthMiddleware(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	secret := "test-jwt-secret-xyz"

	dummyHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "protected-resource")
	}

	t.Run("fails missing bearer token", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodGet, "/protected")

		mw := middleware.Auth(secret, rdb)
		handler := mw(dummyHandler)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("fails invalid token signature", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodGet, "/protected")
		c.Request().Header.Set("Authorization", "Bearer invalid-signature-token")

		mw := middleware.Auth(secret, rdb)
		handler := mw(dummyHandler)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("fails when token is blacklisted in redis", func(t *testing.T) {
		tok := generateTestJWT(secret, "user-123", "", "jti-revoked-999", time.Hour)
		_ = mr.Set("auth:blacklist:jti-revoked-999", "1")

		c, rec := setupEchoTest(http.MethodGet, "/protected")
		c.Request().Header.Set("Authorization", "Bearer "+tok)

		mw := middleware.Auth(secret, rdb)
		handler := mw(dummyHandler)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("success valid token sets user_id in context", func(t *testing.T) {
		tok := generateTestJWT(secret, "user-123", "", "jti-valid-111", time.Hour)

		c, rec := setupEchoTest(http.MethodGet, "/protected")
		c.Request().Header.Set("Authorization", "Bearer "+tok)

		mw := middleware.Auth(secret, rdb)
		handler := mw(func(ctx echo.Context) error {
			assert.Equal(t, "user-123", ctx.Get("user_id"))
			return ctx.String(http.StatusOK, "ok")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestFailureRateLimitMiddleware(t *testing.T) {
	t.Run("passes through cleanly when limiter is nil", func(t *testing.T) {
		c, rec := setupEchoTest(http.MethodPost, "/login")

		mw := middleware.FailureRateLimit(nil, 3, 15*time.Minute, 30*time.Minute)
		handler := mw(func(ctx echo.Context) error {
			return ctx.String(http.StatusOK, "ok")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
