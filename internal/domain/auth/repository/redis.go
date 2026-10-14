package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gdctest/internal/domain/auth/model"

	"github.com/redis/go-redis/v9"
)

type RedisAuthRepository struct {
	rdb *redis.Client
}

func NewRedisAuthRepository(rdb *redis.Client) *RedisAuthRepository {
	return &RedisAuthRepository{rdb: rdb}
}

func (r *RedisAuthRepository) SaveRegistrationSession(ctx context.Context, email string, session *model.RegistrationSession, ttl time.Duration) error {
	if r.rdb == nil {
		return errors.New("redis not connected")
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("auth:reg:%s", email)
	return r.rdb.Set(ctx, key, string(data), ttl).Err()
}

func (r *RedisAuthRepository) GetRegistrationSession(ctx context.Context, email string) (*model.RegistrationSession, error) {
	if r.rdb == nil {
		return nil, errors.New("redis not connected")
	}

	key := fmt.Sprintf("auth:reg:%s", email)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var session model.RegistrationSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RedisAuthRepository) DeleteRegistrationSession(ctx context.Context, email string) error {
	if r.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("auth:reg:%s", email)
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisAuthRepository) SaveResetPasswordSession(ctx context.Context, email string, session *model.RegistrationSession, ttl time.Duration) error {
	if r.rdb == nil {
		return errors.New("redis not connected")
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("auth:reset:%s", email)
	return r.rdb.Set(ctx, key, string(data), ttl).Err()
}

func (r *RedisAuthRepository) GetResetPasswordSession(ctx context.Context, email string) (*model.RegistrationSession, error) {
	if r.rdb == nil {
		return nil, errors.New("redis not connected")
	}

	key := fmt.Sprintf("auth:reset:%s", email)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var session model.RegistrationSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RedisAuthRepository) DeleteResetPasswordSession(ctx context.Context, email string) error {
	if r.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("auth:reset:%s", email)
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisAuthRepository) ClearAllUserSessions(ctx context.Context, email string) error {
	if r.rdb == nil {
		return nil
	}
	regKey := fmt.Sprintf("auth:reg:%s", email)
	resetKey := fmt.Sprintf("auth:reset:%s", email)
	countKey := fmt.Sprintf("ratelimit:count:%s", email)
	return r.rdb.Del(ctx, regKey, resetKey, countKey).Err()
}

func (r *RedisAuthRepository) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	if r.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("auth:blacklist:%s", jti)
	return r.rdb.Set(ctx, key, "1", ttl).Err()
}

func (r *RedisAuthRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	if r.rdb == nil {
		return false, nil
	}
	key := fmt.Sprintf("auth:blacklist:%s", jti)
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
