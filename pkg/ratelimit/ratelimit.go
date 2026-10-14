package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	MaxAttempts int
	Window      time.Duration
	BanDuration time.Duration
}

type Result struct {
	Allowed   bool
	Remaining int
	BanTTL    time.Duration
}

type Limiter struct {
	rdb *redis.Client
}

func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{rdb: rdb}
}

// Check evaluates whether the key is allowed or banned
func (l *Limiter) Check(ctx context.Context, key string, cfg Config) (*Result, error) {
	if l.rdb == nil {
		return &Result{Allowed: true, Remaining: cfg.MaxAttempts}, nil
	}

	banKey := fmt.Sprintf("ratelimit:ban:%s", key)
	countKey := fmt.Sprintf("ratelimit:count:%s", key)

	// 1. Cek apakah sedang dalam masa ban
	banTTL, err := l.rdb.TTL(ctx, banKey).Result()
	if err == nil && banTTL > 0 {
		return &Result{
			Allowed:   false,
			Remaining: 0,
			BanTTL:    banTTL,
		}, nil
	}

	// 2. Increment request count
	count, err := l.rdb.Incr(ctx, countKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to incr ratelimit count: %w", err)
	}

	// Set window expire jika key baru pertama kali dibuat
	if count == 1 {
		_ = l.rdb.Expire(ctx, countKey, cfg.Window).Err()
	}

	// 3. Cek apakah melebihi threshold
	if int(count) > cfg.MaxAttempts {
		// Pasang Ban
		_ = l.rdb.Set(ctx, banKey, "1", cfg.BanDuration).Err()
		_ = l.rdb.Del(ctx, countKey).Err() // reset count

		return &Result{
			Allowed:   false,
			Remaining: 0,
			BanTTL:    cfg.BanDuration,
		}, nil
	}

	return &Result{
		Allowed:   true,
		Remaining: cfg.MaxAttempts - int(count),
		BanTTL:    0,
	}, nil
}

// RecordFailure increments the failure count for path + real IP and bans if threshold exceeded
func (l *Limiter) RecordFailure(ctx context.Context, path, ip string, cfg Config) (*Result, error) {
	if l.rdb == nil {
		return &Result{Allowed: true, Remaining: cfg.MaxAttempts}, nil
	}

	key := fmt.Sprintf("fail:%s:%s", path, ip)
	return l.Check(ctx, key, cfg)
}

// IsBanned checks if a path + real IP is currently banned
func (l *Limiter) IsBanned(ctx context.Context, path, ip string) (bool, time.Duration, error) {
	if l.rdb == nil {
		return false, 0, nil
	}
	key := fmt.Sprintf("fail:%s:%s", path, ip)
	banKey := fmt.Sprintf("ratelimit:ban:%s", key)

	banTTL, err := l.rdb.TTL(ctx, banKey).Result()
	if err == nil && banTTL > 0 {
		return true, banTTL, nil
	}
	return false, 0, nil
}
