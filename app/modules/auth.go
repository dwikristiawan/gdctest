package modules

import (
	"os"
	"time"

	"gdctest/internal/domain/auth/controller"
	"gdctest/internal/domain/auth/repository"
	"gdctest/internal/domain/auth/service"
	"gdctest/internal/middleware"
	"gdctest/pkg/infra"
	"gdctest/pkg/ratelimit"
)

func init() {
	Register(func(ctx *Context) error {
		db, _ := infra.GetDB(ctx.Infra)
		rdb, _ := infra.GetRedis(ctx.Infra)
		mailer, _ := infra.GetMailer(ctx.Infra)

		limiter := ratelimit.NewLimiter(rdb)
		repo := repository.NewAuthRepository(db, rdb)

		baseURL := os.Getenv("APP_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "super-secret-key-12345"
		}

		svc := service.NewAuthService(repo, mailer, limiter, service.AuthServiceConfig{
			BaseURL:   baseURL,
			JWTSecret: jwtSecret,
		})
		ctrl := controller.NewAuthController(svc, limiter)

		// Failure Rate Limit: max 3 kegagalan per path + real IP dalam 15 menit -> ban 30 menit
		failLimitMw := middleware.FailureRateLimit(limiter, 3, 15*time.Minute, 30*time.Minute)

		v1 := ctx.Echo.Group("/v1.0/auth")
		v1.POST("/register", ctrl.Register, failLimitMw)
		v1.POST("/set-password", ctrl.SetPassword, failLimitMw)
		v1.POST("/forgot-password", ctrl.ForgotPassword, failLimitMw)
		v1.POST("/reset-password", ctrl.ResetPassword, failLimitMw)
		v1.POST("/login", ctrl.Login, failLimitMw)
		v1.POST("/logout", ctrl.Logout, ctx.AuthMiddleware)
		v1.GET("/users", ctrl.ListUsers, ctx.AuthMiddleware)

		ctx.RegisterService("auth_service", svc)
		return nil
	})
}
