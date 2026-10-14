package middleware

import (
	"net/http"
	"time"

	"gdctest/pkg/apperror"
	"gdctest/pkg/echo"
	"gdctest/pkg/httputil"
	"gdctest/pkg/ratelimit"

	echoPkg "github.com/labstack/echo/v4"
)

// FailureRateLimit tracks response >= 400 per path and per real IP
func FailureRateLimit(limiter *ratelimit.Limiter, maxAttempts int, window, banDuration time.Duration) echoPkg.MiddlewareFunc {
	return func(next echoPkg.HandlerFunc) echoPkg.HandlerFunc {
		return func(c echoPkg.Context) error {
			if limiter == nil {
				return next(c)
			}

			path := c.Path()
			ip := httputil.GetRealIP(c)

			// 1. Cek apakah IP sedang di-ban untuk path ini
			isBanned, banTTL, err := limiter.IsBanned(c.Request().Context(), path, ip)
			if err == nil && isBanned {
				return echo.Respond(c, nil, apperror.New(http.StatusTooManyRequests, "too many failed attempts on this endpoint. banned for "+banTTL.Round(time.Second).String()))
			}

			// 2. Jalankan handler
			execErr := next(c)

			// 3. Evaluasi status response (apakah >= 400)
			status := c.Response().Status
			if status >= 400 {
				_, _ = limiter.RecordFailure(c.Request().Context(), path, ip, ratelimit.Config{
					MaxAttempts: maxAttempts,
					Window:      window,
					BanDuration: banDuration,
				})
			}

			return execErr
		}
	}
}
