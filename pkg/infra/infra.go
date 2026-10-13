package infra

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Closer interface {
	Shutdown(ctx context.Context) error
}

type Container struct {
	resources map[string]any
	closers   []Closer
	mu        sync.RWMutex
}

var globalContainer = &Container{
	resources: make(map[string]any),
}

func (c *Container) Set(name string, res any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resources[name] = res
	if closer, ok := res.(Closer); ok {
		c.closers = append(c.closers, closer)
	}
}

func (c *Container) Get(name string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, ok := c.resources[name]
	return res, ok
}

func (c *Container) ShutdownAll(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, closer := range c.closers {
		if err := closer.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

func Get[T any](name string) (T, bool) {
	res, ok := globalContainer.Get(name)
	if !ok {
		var zero T
		return zero, false
	}
	val, ok := res.(T)
	return val, ok
}

type InitializerFunc func(ctx context.Context, c *Container) error

var (
	initializers = make(map[string]InitializerFunc)
	initMu       sync.Mutex
)

// Register mendaftarkan initializer infra (dipanggil via init() di tiap file infra)
func Register(name string, fn InitializerFunc) {
	initMu.Lock()
	defer initMu.Unlock()
	initializers[name] = fn
}

// InitAll menjalankan semua initializer terdaftar secara async/paralel dengan errgroup
func InitAll(ctx context.Context) (*Container, error) {
	initMu.Lock()
	defer initMu.Unlock()

	g, gCtx := errgroup.WithContext(ctx)

	for name, fn := range initializers {
		initFn := fn
		resName := name
		g.Go(func() error {
			if err := initFn(gCtx, globalContainer); err != nil {
				slog.Error("failed to initialize infrastructure", "infra", resName, "error", err)
				return fmt.Errorf("infra %s failed to initialize: %w", resName, err)
			}
			slog.Info("infrastructure ready", "infra", resName)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		_ = globalContainer.ShutdownAll(context.Background())
		return nil, err
	}

	return globalContainer, nil
}
