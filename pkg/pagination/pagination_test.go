package pagination_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gdctest/pkg/pagination"
	"gdctest/pkg/squtil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type sampleUser struct {
	ID       string
	Email    string
	Username string
}

func TestFromEcho(t *testing.T) {
	e := echo.New()

	t.Run("defaults when query parameters empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		c := e.NewContext(req, httptest.NewRecorder())

		p := pagination.FromEcho(c)
		assert.Equal(t, 1, p.Page)
		assert.Equal(t, 10, p.Limit)
		assert.Empty(t, p.Order)
	})

	t.Run("parses valid params and caps max limit to 100", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test?page=3&limit=500&order=created_at:asc", nil)
		c := e.NewContext(req, httptest.NewRecorder())

		p := pagination.FromEcho(c)
		assert.Equal(t, 3, p.Page)
		assert.Equal(t, 100, p.Limit) // Capped at 100
		assert.Equal(t, "created_at:asc", p.Order)
	})
}

func TestFetch_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	ctx := context.Background()

	// Base query dengan WHERE
	baseQuery := squtil.PG.
		Select("id", "email", "username").
		From("users").
		Where(squirrel.Eq{"email": "test@example.com"})

	req := pagination.Request{
		Page:  1,
		Limit: 2,
		Order: "created_at:desc",
	}

	allowedSort := map[string]bool{"created_at": true, "username": true}

	// 1. Expected Count Query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(SELECT id, email, username FROM users WHERE email = \$1\) AS count_subquery`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// 2. Expected List Data Query
	mock.ExpectQuery(`SELECT id, email, username FROM users WHERE email = \$1 ORDER BY created_at DESC, id DESC LIMIT 2 OFFSET 0`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "username"}).
			AddRow("u1", "test@example.com", "user1").
			AddRow("u2", "test@example.com", "user2"))

	items, meta, err := pagination.Fetch(
		ctx,
		db,
		baseQuery,
		req,
		allowedSort,
		[]string{"created_at DESC"},
		func(rows *sql.Rows) (sampleUser, error) {
			var u sampleUser
			err := rows.Scan(&u.ID, &u.Email, &u.Username)
			return u, err
		},
	)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "u1", items[0].ID)
	assert.Equal(t, "u2", items[1].ID)

	assert.NotNil(t, meta)
	assert.Equal(t, int64(5), meta.TotalRows)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 2, meta.Limit)
	assert.Equal(t, 3, meta.TotalPages) // Ceil(5 / 2) = 3
	assert.True(t, meta.HasNext)
	assert.False(t, meta.HasPrev)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFetch_ErrorHandling(t *testing.T) {
	t.Run("fails when count query fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		baseQuery := squtil.PG.Select("id").From("users")

		mock.ExpectQuery(`SELECT COUNT\(\*\)`).
			WillReturnError(errors.New("count query failure"))
		mock.ExpectQuery(`SELECT id FROM users`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("u1"))

		_, _, err = pagination.Fetch(
			context.Background(),
			db,
			baseQuery,
			pagination.Request{Page: 1, Limit: 10},
			nil,
			nil,
			func(rows *sql.Rows) (string, error) {
				var id string
				return id, rows.Scan(&id)
			},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "count query failure")
	})

	t.Run("fails when list query fails", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		mock.MatchExpectationsInOrder(false)

		baseQuery := squtil.PG.Select("id").From("users")

		mock.ExpectQuery(`SELECT COUNT\(\*\)`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`SELECT id FROM users`).
			WillReturnError(errors.New("list query failure"))

		_, _, err = pagination.Fetch(
			context.Background(),
			db,
			baseQuery,
			pagination.Request{Page: 1, Limit: 10},
			nil,
			nil,
			func(rows *sql.Rows) (string, error) {
				var id string
				return id, rows.Scan(&id)
			},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list query failure")
	})
}
