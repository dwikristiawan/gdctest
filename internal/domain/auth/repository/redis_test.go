package repository_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/repository"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisAuthRepository_RegistrationSession(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := repository.NewRedisAuthRepository(db)
	ctx := context.Background()

	email := "test@example.com"
	session := &model.RegistrationSession{Token: "tok-123", IP: "127.0.0.1"}
	data, _ := json.Marshal(session)
	key := "auth:reg:" + email

	t.Run("save registration session", func(t *testing.T) {
		mock.ExpectSet(key, string(data), 30*time.Minute).SetVal("OK")
		err := repo.SaveRegistrationSession(ctx, email, session, 30*time.Minute)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get registration session", func(t *testing.T) {
		mock.ExpectGet(key).SetVal(string(data))
		res, err := repo.GetRegistrationSession(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "tok-123", res.Token)
		assert.Equal(t, "127.0.0.1", res.IP)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete registration session", func(t *testing.T) {
		mock.ExpectDel(key).SetVal(1)
		err := repo.DeleteRegistrationSession(ctx, email)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisAuthRepository_ResetPasswordSession(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := repository.NewRedisAuthRepository(db)
	ctx := context.Background()

	email := "reset@example.com"
	session := &model.RegistrationSession{Token: "reset-tok-456", IP: "192.168.1.1"}
	data, _ := json.Marshal(session)
	key := "auth:reset:" + email

	t.Run("save and get reset session", func(t *testing.T) {
		mock.ExpectSet(key, string(data), 15*time.Minute).SetVal("OK")
		err := repo.SaveResetPasswordSession(ctx, email, session, 15*time.Minute)
		assert.NoError(t, err)

		mock.ExpectGet(key).SetVal(string(data))
		res, err := repo.GetResetPasswordSession(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "reset-tok-456", res.Token)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete reset session", func(t *testing.T) {
		mock.ExpectDel(key).SetVal(1)
		err := repo.DeleteResetPasswordSession(ctx, email)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisAuthRepository_ClearAllUserSessions(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := repository.NewRedisAuthRepository(db)
	ctx := context.Background()

	email := "clear@example.com"
	regKey := "auth:reg:" + email
	resetKey := "auth:reset:" + email
	countKey := "ratelimit:count:" + email

	mock.ExpectDel(regKey, resetKey, countKey).SetVal(3)

	err := repo.ClearAllUserSessions(ctx, email)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisAuthRepository_BlacklistToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	repo := repository.NewRedisAuthRepository(db)
	ctx := context.Background()

	jti := "uuid-jti-token-999"
	key := "auth:blacklist:" + jti

	t.Run("blacklist token and verify exists", func(t *testing.T) {
		mock.ExpectSet(key, "1", 1*time.Hour).SetVal("OK")
		err := repo.BlacklistToken(ctx, jti, 1*time.Hour)
		assert.NoError(t, err)

		mock.ExpectExists(key).SetVal(1)
		isBlacklisted, err := repo.IsTokenBlacklisted(ctx, jti)
		assert.NoError(t, err)
		assert.True(t, isBlacklisted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
