package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gdctest/pkg/logger"

	"github.com/labstack/echo/v4"
)

func BenchmarkRequestLogger(b *testing.B) {
	logger.Init()
	e := echo.New()
	handler := RequestLogger()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"status":"ping"}`))
	req.Header.Set("Content-Type", "application/json")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = handler(c)
	}
}
