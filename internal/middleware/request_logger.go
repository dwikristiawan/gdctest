package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"gdctest/pkg/logger"

	"github.com/labstack/echo/v4"
)

var bufPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

type loggingResponseWriter struct {
	responseWriter http.ResponseWriter
	captureBuf     *bytes.Buffer
}

func (w *loggingResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if w.captureBuf != nil {
		_, _ = w.captureBuf.Write(b)
	}
	return w.responseWriter.Write(b)
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.responseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Flush() {
	if flusher, ok := w.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

var lrwPool = sync.Pool{
	New: func() any {
		return &loggingResponseWriter{}
	},
}

const maxCaptureBytes = 32 * 1024 // ponytail: 32KB cap prevents OOM on large multipart/file uploads

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			var reqBody []byte
			if req.Body != nil {
				reqBuf := bufPool.Get().(*bytes.Buffer)
				reqBuf.Reset()
				_, _ = io.CopyN(reqBuf, req.Body, maxCaptureBytes)
				reqBody = reqBuf.Bytes()
				if len(reqBody) > 0 {
					req.Body = io.NopCloser(bytes.NewReader(reqBody))
				}
				defer bufPool.Put(reqBuf)
			}

			resBuf := bufPool.Get().(*bytes.Buffer)
			resBuf.Reset()
			defer bufPool.Put(resBuf)

			origWriter := c.Response().Writer
			lrw := lrwPool.Get().(*loggingResponseWriter)
			lrw.responseWriter = origWriter
			lrw.captureBuf = resBuf
			c.Response().Writer = lrw

			err := next(c)

			c.Response().Writer = origWriter
			lrw.responseWriter = nil
			lrw.captureBuf = nil
			lrwPool.Put(lrw)

			latencyMs := time.Since(start).Milliseconds()
			reqContentType := req.Header.Get("Content-Type")
			resContentType := c.Response().Header().Get("Content-Type")

			reqBodyFormatted := logger.FormatPayload(reqBody, reqContentType)
			resBodyFormatted := logger.FormatPayload(resBuf.Bytes(), resContentType)

			logAttrs := []slog.Attr{
				slog.String("url", req.URL.Path),
				slog.String("method", req.Method),
				slog.String("route", c.Path()),
				slog.String("path", req.URL.Path),
				slog.Int("status_code", c.Response().Status),
				slog.Int64("latency_ms", latencyMs),
				slog.String("client_ip", c.RealIP()),
				slog.Group("request",
					slog.Any("headers", req.Header),
					slog.Any("body", reqBodyFormatted),
				),
				slog.Group("response",
					slog.Any("headers", c.Response().Header()),
					slog.Any("body", resBodyFormatted),
				),
			}

			if realErr := c.Get("real_error"); realErr != nil {
				logAttrs = append(logAttrs, slog.Any("real_error", realErr))
			}

			level := slog.LevelInfo
			if c.Response().Status >= http.StatusInternalServerError {
				level = slog.LevelError
			} else if c.Response().Status >= http.StatusBadRequest {
				level = slog.LevelWarn
			}

			logger.LogAttrsDirect(req.Context(), level, "client request", logAttrs...)

			return err
		}
	}
}
