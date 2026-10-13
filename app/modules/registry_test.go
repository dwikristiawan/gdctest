package modules

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestModules_RegistrationAndBoot(t *testing.T) {
	ResetRegistry()

	var boot1Called, boot2Called bool

	Register(func(ctx *Context) error {
		boot1Called = true
		ctx.RegisterService("mock1", "service1")
		return nil
	})

	Register(func(ctx *Context) error {
		boot2Called = true
		svc, ok := ctx.GetService("mock1")
		if !ok || svc != "service1" {
			t.Errorf("expected shared service mock1 to be accessible")
		}
		return nil
	})

	e := echo.New()
	ctx := NewContext(e.Group(""), nil, nil)

	if err := BootAll(ctx); err != nil {
		t.Fatalf("BootAll failed: %v", err)
	}

	if !boot1Called || !boot2Called {
		t.Errorf("expected both modules to be booted: boot1=%v, boot2=%v", boot1Called, boot2Called)
	}
}
