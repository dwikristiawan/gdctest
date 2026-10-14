package httputil

import (
	"net"
	"strings"

	"github.com/labstack/echo/v4"
)

// GetRealIP extracts client real IP from headers with fallback to remote address
func GetRealIP(c echo.Context) string {
	// 1. X-Forwarded-For
	if xff := c.Request().Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 2. X-Real-IP
	if xri := c.Request().Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 3. Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err == nil {
		return host
	}
	return c.Request().RemoteAddr
}
