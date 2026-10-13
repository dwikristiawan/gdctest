package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gdctest/pkg/ctxkey"
)

type Logger struct {
	base   *slog.Logger
	buf    [32]slog.Attr
	fields []slog.Attr
}

type Config struct {
	Level  string
	Format string
	Output io.Writer
}

var defaultLogger *Logger

func Init(cfgs ...Config) {
	var cfg Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	if cfg.Level == "" {
		cfg.Level = os.Getenv("LOG_LEVEL")
	}
	if cfg.Format == "" {
		cfg.Format = os.Getenv("LOG_FORMAT")
	}

	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
					return slog.String("line", src.File+":"+strconv.Itoa(src.Line))
				}
			}
			if a.Key == slog.MessageKey && a.Value.String() == "" {
				return slog.Attr{}
			}
			return a
		},
	}

	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	base := slog.New(handler)
	slog.SetDefault(base)
	defaultLogger = &Logger{base: base}
}

func Default() *Logger {
	if defaultLogger == nil {
		return &Logger{base: slog.Default()}
	}
	return defaultLogger
}

func FromContext(ctx context.Context) *Logger {
	base := slog.Default()
	if defaultLogger != nil && defaultLogger.base != nil {
		base = defaultLogger.base
	}
	l := &Logger{base: base}
	l.fields = l.buf[:0]
	if ctx == nil {
		return l
	}
	if id, ok := ctx.Value(ctxkey.RequestID).(string); ok && id != "" {
		l.fields = append(l.fields, slog.String("request_id", id))
	}
	return l
}

// LogAttrsDirect emits a log record directly to the handler with 0 struct allocations.
func LogAttrsDirect(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	base := slog.Default()
	if defaultLogger != nil && defaultLogger.base != nil {
		base = defaultLogger.base
	}
	if !base.Enabled(ctx, level) {
		return
	}

	var pcs [8]uintptr
	n := runtime.Callers(2, pcs[:])
	var callerPC uintptr
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i] - 1)
		if fn != nil {
			name := fn.Name()
			if !strings.Contains(name, "/pkg/logger.") {
				callerPC = pcs[i]
				break
			}
		}
	}

	r := slog.NewRecord(time.Now(), level, msg, callerPC)
	if ctx != nil {
		if id, ok := ctx.Value(ctxkey.RequestID).(string); ok && id != "" {
			r.AddAttrs(slog.String("request_id", id))
		}
	}
	r.AddAttrs(attrs...)
	_ = base.Handler().Handle(ctx, r)
}

func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	c := l.clone()
	c.fields = append(c.fields, attrs...)
	return c
}

func (l *Logger) WithData(data any) *Logger {
	c := l.clone()
	if b, ok := data.([]byte); ok {
		data = FormatPayload(b, "")
	}
	c.fields = append(c.fields, slog.Any("data", data))
	return c
}

func FormatPayload(body []byte, contentType string) any {
	if len(body) == 0 {
		return nil
	}

	ct := contentType
	if idx := strings.IndexByte(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}
	ct = strings.TrimSpace(ct)

	if ct == "" {
		sniffLen := len(body)
		if sniffLen > 512 {
			sniffLen = 512
		}
		ct = http.DetectContentType(body[:sniffLen])
		if idx := strings.IndexByte(ct, ';'); idx >= 0 {
			ct = ct[:idx]
		}
		ct = strings.TrimSpace(ct)
	}

	if strings.Contains(strings.ToLower(ct), "json") {
		if json.Valid(body) {
			return json.RawMessage(body)
		}
	}

	return fmt.Sprintf("[%s, %s]", ct, formatBytes(len(body)))
}

func formatBytes(b int) string {
	const unit = 1024
	if b < unit {
		return strconv.Itoa(b) + " B"
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (l *Logger) WithError(err error) *Logger {
	c := l.clone()
	if err != nil {
		c.fields = append(c.fields, slog.String("error", err.Error()))
	}
	return c
}

func (l *Logger) WithField(key string, val any) *Logger {
	c := l.clone()
	c.fields = append(c.fields, slog.Any(key, val))
	return c
}

func (l *Logger) WithFields(fields map[string]any) *Logger {
	c := l.clone()
	for k, v := range fields {
		c.fields = append(c.fields, slog.Any(k, v))
	}
	return c
}

func (l *Logger) Info(input any, args ...any) {
	l.dispatchLog(slog.LevelInfo, input, args...)
}

func (l *Logger) Warn(input any, args ...any) {
	l.dispatchLog(slog.LevelWarn, input, args...)
}

func (l *Logger) Debug(input any, args ...any) {
	l.dispatchLog(slog.LevelDebug, input, args...)
}

func (l *Logger) Error(input any, args ...any) {
	if err, ok := input.(error); ok {
		if err != nil {
			l.fields = append(l.fields, slog.String("error", err.Error()))
		}
		if len(args) == 1 {
			l.dispatchLog(slog.LevelError, args[0])
			return
		}
		l.dispatchLog(slog.LevelError, "", args...)
		return
	}
	l.dispatchLog(slog.LevelError, input, args...)
}

func (l *Logger) dispatchLog(level slog.Level, input any, args ...any) {
	if s, ok := input.(string); ok {
		l.log(level, s, args...)
		return
	}
	if b, ok := input.([]byte); ok {
		input = FormatPayload(b, "")
	}
	l.fields = append(l.fields, slog.Any("data", input))
	l.log(level, "", args...)
}

func (l *Logger) clone() *Logger {
	c := &Logger{base: l.base}
	if len(l.fields) <= len(c.buf) {
		c.fields = c.buf[:len(l.fields)]
		copy(c.fields, l.fields)
	} else {
		c.fields = make([]slog.Attr, len(l.fields), len(l.fields)+4)
		copy(c.fields, l.fields)
	}
	return c
}

func (l *Logger) log(level slog.Level, msg string, args ...any) {
	ctx := context.Background()
	if !l.base.Enabled(ctx, level) {
		return
	}

	var pcs [8]uintptr
	n := runtime.Callers(2, pcs[:])
	var callerPC uintptr
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i] - 1)
		if fn != nil {
			name := fn.Name()
			if !strings.Contains(name, "/pkg/logger.") && !strings.Contains(name, "/pkg/logger.(*") {
				callerPC = pcs[i]
				break
			}
		}
	}

	r := slog.NewRecord(time.Now(), level, msg, callerPC)
	r.AddAttrs(l.fields...)
	r.Add(args...)
	_ = l.base.Handler().Handle(ctx, r)
}

func (l *Logger) Slog() *slog.Logger {
	args := make([]any, 0, len(l.fields)*2)
	for _, f := range l.fields {
		args = append(args, f.Key, f.Value.Any())
	}
	return l.base.With(args...)
}
