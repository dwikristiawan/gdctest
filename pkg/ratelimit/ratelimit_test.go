package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"gdctest/pkg/ratelimit"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestLimiter_Check(t *testing.T) {
	ctx := context.Background()
	cfg := ratelimit.Config{
		MaxAttempts: 3,
		Window:      15 * time.Minute,
		BanDuration: 30 * time.Minute,
	}

	t.Run("nil redis client allows everything", func(t *testing.T) {
		limiter := ratelimit.NewLimiter(nil)
		res, err := limiter.Check(ctx, "test-key", cfg)
		assert.NoError(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, 3, res.Remaining)
	})

	t.Run("first attempt allowed and sets window expiry", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		limiter := ratelimit.NewLimiter(rdb)

		mock.ExpectTTL("ratelimit:ban:user@example.com").SetVal(-2 * time.Second) // no ban
		mock.ExpectIncr("ratelimit:count:user@example.com").SetVal(1)
		mock.ExpectExpire("ratelimit:count:user@example.com", 15*time.Minute).SetVal(true)

		res, err := limiter.Check(ctx, "user@example.com", cfg)
		assert.NoError(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, 2, res.Remaining)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exceeding max attempts triggers ban and deletes count key", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		limiter := ratelimit.NewLimiter(rdb)

		mock.ExpectTTL("ratelimit:ban:user@example.com").SetVal(-2 * time.Second)
		mock.ExpectIncr("ratelimit:count:user@example.com").SetVal(4) // 4 > 3
		mock.ExpectSet("ratelimit:ban:user@example.com", "1", 30*time.Minute).SetVal("OK")
		mock.ExpectDel("ratelimit:count:user@example.com").SetVal(1)

		res, err := limiter.Check(ctx, "user@example.com", cfg)
		assert.NoError(t, err)
		assert.False(t, res.Allowed)
		assert.Equal(t, 0, res.Remaining)
		assert.Equal(t, 30*time.Minute, res.BanTTL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("already banned key returns allowed false with remaining TTL", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		limiter := ratelimit.NewLimiter(rdb)

		mock.ExpectTTL("ratelimit:ban:user@example.com").SetVal(20 * time.Minute)

		res, err := limiter.Check(ctx, "user@example.com", cfg)
		assert.NoError(t, err)
		assert.False(t, res.Allowed)
		assert.Equal(t, 20*time.Minute, res.BanTTL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLimiter_RecordFailureAndIsBanned(t *testing.T) {
	ctx := context.Background()
	cfg := ratelimit.Config{
		MaxAttempts: 3,
		Window:      15 * time.Minute,
		BanDuration: 30 * time.Minute,
	}

	t.Run("record failure delegates to check with path and IP key", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		limiter := ratelimit.NewLimiter(rdb)

		expectedKey := "fail:/api/v1/login:192.168.1.5"
		mock.ExpectTTL("ratelimit:ban:" + expectedKey).SetVal(-2 * time.Second)
		mock.ExpectIncr("ratelimit:count:" + expectedKey).SetVal(1)
		mock.ExpectExpire("ratelimit:count:"+expectedKey, 15*time.Minute).SetVal(true)

		res, err := limiter.RecordFailure(ctx, "/api/v1/login", "192.168.1.5", cfg)
		assert.NoError(t, err)
		assert.True(t, res.Allowed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("is banned returns true when TTL positive", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		limiter := ratelimit.NewLimiter(rdb)

		expectedKey := "fail:/api/v1/login:192.168.1.5"
		mock.ExpectTTL("ratelimit:ban:" + expectedKey).SetVal(10 * time.Minute)

		banned, ttl, err := limiter.IsBanned(ctx, "/api/v1/login", "192.168.1.5")
		assert.NoError(t, err)
		assert.True(t, banned)
		assert.Equal(t, 10*time.Minute, ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
