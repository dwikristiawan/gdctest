package modules

import (
	"gdctest/internal/domain/health/controller"
	"gdctest/internal/domain/health/service"
)

func init() {
	Register(func(ctx *Context) error {
		svc := service.NewHealthService(ctx.DB, ctx.Redis)
		ctrl := controller.NewHealthController(svc)

		v1 := ctx.Echo.Group("/v1.0/healthcheck")
		v1.GET("/liveness", ctrl.Liveness)
		v1.GET("/readiness", ctrl.Readiness)

		ctx.RegisterService("health", svc)
		return nil
	})
}
