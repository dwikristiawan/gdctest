package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gdctest/pkg/httputil"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetRealIP(t *testing.T) {
	e := echo.New()

	t.Run("extracts first IP from X-Forwarded-For multiple proxies", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.195, 198.51.100.1, 10.0.0.1")
		req.RemoteAddr = "10.0.0.2:12345"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ip := httputil.GetRealIP(c)
		assert.Equal(t, "203.0.113.195", ip)
	})

	t.Run("extracts from X-Real-IP when X-Forwarded-For is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "198.51.100.50")
		req.RemoteAddr = "10.0.0.2:12345"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ip := httputil.GetRealIP(c)
		assert.Equal(t, "198.51.100.50", ip)
	})

	t.Run("fallback to RemoteAddr stripping port", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.100:54321"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ip := httputil.GetRealIP(c)
		assert.Equal(t, "192.168.1.100", ip)
	})

	t.Run("fallback to raw RemoteAddr when no port present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.100"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ip := httputil.GetRealIP(c)
		assert.Equal(t, "192.168.1.100", ip)
	})
}
