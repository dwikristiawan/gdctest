package modules

import (
	"database/sql"
	"fmt"
	"sync"

	"gdctest/app/config"
	"gdctest/pkg/database"

	"github.com/labstack/echo/v4"
)

type Context struct {
	Echo     *echo.Group
	DB       *sql.DB
	Redis    *database.RedisClient
	Config   *config.Config
	services map[string]any
	mu       sync.RWMutex
}

func NewContext(echoGroup *echo.Group, db *sql.DB, redisClient *database.RedisClient, cfg *config.Config) *Context {
	return &Context{
		Echo:     echoGroup,
		DB:       db,
		Redis:    redisClient,
		Config:   cfg,
		services: make(map[string]any),
	}
}

// RegisterService allows a module to expose a shared service to other modules safely
func (c *Context) RegisterService(name string, svc any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = svc
}

// GetService retrieves a shared service by name
func (c *Context) GetService(name string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	svc, ok := c.services[name]
	return svc, ok
}

type ModuleFunc func(ctx *Context) error

var (
	registry []ModuleFunc
	regMu    sync.Mutex
)

// Register registers a module bootstrap function (typically called inside init())
func Register(fn ModuleFunc) {
	regMu.Lock()
	defer regMu.Unlock()
	registry = append(registry, fn)
}

// BootAll initializes and mounts all registered modules onto the given Context
func BootAll(ctx *Context) error {
	regMu.Lock()
	defer regMu.Unlock()

	for _, fn := range registry {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("module boot failed: %w", err)
		}
	}
	return nil
}

// ResetRegistry clears the registry (useful for unit tests)
func ResetRegistry() {
	regMu.Lock()
	defer regMu.Unlock()
	registry = nil
}
