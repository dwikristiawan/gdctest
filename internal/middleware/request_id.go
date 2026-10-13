package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"gdctest/pkg/ctxkey"

	"github.com/labstack/echo/v4"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := c.Request().Header.Get("X-Request-ID")
			if rid == "" {
				var b [16]byte
				_, _ = rand.Read(b[:])
				rid = hex.EncodeToString(b[:])
			}
			c.Response().Header().Set("X-Request-ID", rid)
			c.Set("request_id", rid)

			ctx := context.WithValue(c.Request().Context(), ctxkey.RequestID, rid)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
