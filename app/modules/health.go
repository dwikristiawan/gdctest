package modules

import (
	"gdctest/internal/domain/health/controller"
	"gdctest/internal/domain/health/service"
)

func init() {
	Register(func(ctx *Context) error {
		svc := service.NewHealthService()
		ctrl := controller.NewHealthController(svc)

		ctx.Echo.GET("/liveness", ctrl.Liveness)
		ctx.RegisterService("health", svc)
		return nil
	})
}
