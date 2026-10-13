package controller

import (
	"gdctest/internal/domain/health/service"
	pkgEcho "gdctest/pkg/echo"

	"github.com/labstack/echo/v4"
)

type HealthControllerInterface interface {
	Liveness(c echo.Context) error
}

type HealthController struct {
	svc service.HealthServiceInterface
}

func NewHealthController(svc service.HealthServiceInterface) HealthControllerInterface {
	return &HealthController{svc: svc}
}

func (h *HealthController) Liveness(c echo.Context) error {
	data, err := h.svc.Liveness()
	return pkgEcho.Respond(c, data, err)
}
