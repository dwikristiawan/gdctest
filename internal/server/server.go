package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gdctest/app/config"
	"gdctest/app/modules"
	"gdctest/internal/middleware"
	"gdctest/pkg/infra"

	"github.com/go-playground/validator/v10"
	echoV4 "github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)

type echoValidator struct{ v *validator.Validate }

func (ev *echoValidator) Validate(i interface{}) error { return ev.v.Struct(i) }

func Run(cfg *config.Config) {
	// 1. Initialize all registered infrastructure concurrently
	infraContainer, err := infra.InitAll(context.Background())
	if err != nil {
		slog.Error("infrastructure initialization failed, aborting server startup", "error", err)
		os.Exit(1)
	}

	// 2. Initialize Echo HTTP Server & Middlewares
	e := echoV4.New()
	e.HideBanner = true
	e.Validator = &echoValidator{v: validator.New()}

	e.Use(echoMw.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS())

	// 3. Boot all modules registered in app/modules
	modCtx := modules.NewContext(e.Group("/api"), infraContainer, cfg)
	if err := modules.BootAll(modCtx); err != nil {
		slog.Error("failed to boot modules", "error", err)
		os.Exit(1)
	}

	// 4. Start HTTP Server
	go func() {
		slog.Info("server starting", "addr", cfg.Server.Domain)
		if err := e.Start(cfg.Server.Domain); err != nil && err != http.ErrServerClosed {
			slog.Info("server stopped", "reason", err.Error())
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("echo shutdown error", "error", err)
	}
	if err := infraContainer.ShutdownAll(ctx); err != nil {
		slog.Error("infra shutdown error", "error", err)
	}
	slog.Info("server stopped gracefully")
}
