package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     int    `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

type RedisClient struct {
	Client *redis.Client
}

func (r *RedisClient) Shutdown(ctx context.Context) error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

// GetRedis returns typed *redis.Client safely from Container
func GetRedis(c *Container) (*redis.Client, bool) {
	if c == nil {
		return nil, false
	}
	if res, ok := c.Get("redis"); ok {
		if r, ok := res.(*RedisClient); ok {
			return r.Client, true
		}
	}
	return nil, false
}

func init() {
	Register("redis", func(ctx context.Context, c *Container) error {
		var cfg RedisConfig
		if err := envconfig.Process("", &cfg); err != nil {
			return fmt.Errorf("failed to parse redis config: %w", err)
		}

		rdb := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := rdb.Ping(pingCtx).Err(); err != nil {
			_ = rdb.Close()
			return fmt.Errorf("failed to ping redis: %w", err)
		}

		c.Set("redis", &RedisClient{Client: rdb})
		return nil
	})
}
