package repository_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"gdctest/internal/domain/auth/model"
	"gdctest/internal/domain/auth/repository"
	"gdctest/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresUserRepository_FindOne(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresUserRepository(db)
	ctx := context.Background()
	email := "user@example.com"

	t.Run("success found user by email object", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "username", "password_hash", "created_at", "updated_at"}).
			AddRow("uuid-123", email, "user1", "hashed", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, username, password_hash, created_at, updated_at FROM users WHERE email = $1 LIMIT 1`)).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.FindOne(ctx, model.User{Email: email})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "user1", user.Username)
	})

	t.Run("user not found returns nil", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, username, password_hash, created_at, updated_at FROM users WHERE email = $1 LIMIT 1`)).
			WithArgs("notfound@example.com").
			WillReturnError(sql.ErrNoRows)

		user, err := repo.FindOne(ctx, model.User{Email: "notfound@example.com"})
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestPostgresUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresUserRepository(db)
	ctx := context.Background()

	email := "new@example.com"
	username := "newuser"
	hash := "somehash"

	rows := sqlmock.NewRows([]string{"id", "email", "username", "created_at", "updated_at"}).
		AddRow("new-uuid-456", email, username, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (email,username,password_hash) VALUES ($1,$2,$3) RETURNING id, email, username, created_at, updated_at`)).
		WithArgs(email, username, hash).
		WillReturnRows(rows)

	created, err := repo.CreateUser(ctx, email, username, hash)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, email, created.Email)
	assert.Equal(t, username, created.Username)
}

func TestPostgresUserRepository_UpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresUserRepository(db)
	ctx := context.Background()

	email := "user@example.com"
	newHash := "newhash123"

	t.Run("success update user password", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE email = $2`)).
			WithArgs(newHash, email).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateUser(ctx, model.User{Email: email}, model.User{PasswordHash: newHash})
		assert.NoError(t, err)
	})

	t.Run("user not found returns error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE email = $2`)).
			WithArgs(newHash, "ghost@example.com").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateUser(ctx, model.User{Email: "ghost@example.com"}, model.User{PasswordHash: newHash})
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})
}

func TestPostgresUserRepository_FindUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Asynchronous goroutines execute queries concurrently via errgroup
	mock.MatchExpectationsInOrder(false)

	repo := repository.NewPostgresUserRepository(db)
	ctx := context.Background()

	t.Run("success find users with pagination", func(t *testing.T) {
		req := pagination.Request{
			Page:  1,
			Limit: 10,
			Order: "created_at:desc",
		}

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(SELECT id, email, username, created_at, updated_at FROM users\) AS count_subquery`).
			WillReturnRows(countRows)

		dataRows := sqlmock.NewRows([]string{"id", "email", "username", "created_at", "updated_at"}).
			AddRow("u-1", "user1@example.com", "user1", time.Now(), time.Now()).
			AddRow("u-2", "user2@example.com", "user2", time.Now(), time.Now())
		mock.ExpectQuery(`SELECT id, email, username, created_at, updated_at FROM users ORDER BY created_at DESC, id DESC LIMIT 10 OFFSET 0`).
			WillReturnRows(dataRows)

		users, meta, err := repo.FindUsers(ctx, model.User{}, req)
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, int64(2), meta.TotalRows)
		assert.Len(t, users, 2)
		assert.Equal(t, "user1@example.com", users[0].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

