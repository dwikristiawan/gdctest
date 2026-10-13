package service

import (
	"context"
	"time"

	"gdctest/pkg/database"
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
	redis *database.RedisClient
}

func NewHealthService(redis *database.RedisClient) HealthServiceInterface {
	return &HealthService{redis: redis}
}

func (s *HealthService) Liveness() (*HealthResponse, error) {
	return &HealthResponse{Status: "ok"}, nil
}

func (s *HealthService) Readiness(ctx context.Context) (*HealthResponse, error) {
	checks := make(map[string]string)
	status := "ok"

	if s.redis != nil && s.redis.Client != nil {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := s.redis.Client.Ping(pingCtx).Err(); err != nil {
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
