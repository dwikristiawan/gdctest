package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

type HealthServiceInterface interface {
	Liveness() (*HealthResponse, error)
	Readiness(ctx context.Context) (*HealthResponse, error)
}

type HealthService struct {
	db    *sql.DB
	redis *redis.Client
}

func NewHealthService(db *sql.DB, redisClient *redis.Client) HealthServiceInterface {
	return &HealthService{
		db:    db,
		redis: redisClient,
	}
}

func (s *HealthService) Liveness() (*HealthResponse, error) {
	return &HealthResponse{Status: "ok"}, nil
}

func (s *HealthService) Readiness(ctx context.Context) (*HealthResponse, error) {
	checks := make(map[string]string)
	status := "ok"

	// Postgres check
	if s.db != nil {
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := s.db.PingContext(dbCtx); err != nil {
			checks["postgres"] = "down: " + err.Error()
			status = "degraded"
		} else {
			checks["postgres"] = "pong"
		}
	} else {
		checks["postgres"] = "unconfigured"
		status = "degraded"
	}

	// Redis check
	if s.redis != nil {
		redisCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := s.redis.Ping(redisCtx).Err(); err != nil {
			checks["redis"] = "down: " + err.Error()
			status = "degraded"
		} else {
			checks["redis"] = "pong"
		}
	} else {
		checks["redis"] = "unconfigured"
		status = "degraded"
	}

	return &HealthResponse{
		Status: status,
		Checks: checks,
	}, nil
}
