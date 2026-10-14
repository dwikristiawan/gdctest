package repository

import (
	"context"
	"database/sql"
	"errors"

	"gdctest/internal/domain/auth/model"
	"gdctest/pkg/pagination"
	"gdctest/pkg/squtil"

	"github.com/Masterminds/squirrel"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// FindOne finds a single user based on non-empty fields in the model.User filter
func (r *PostgresUserRepository) FindOne(ctx context.Context, filter model.User) (*model.User, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	builder := squtil.PG.
		Select("id", "email", "username", "password_hash", "created_at", "updated_at").
		From("users")

	if filter.ID != "" {
		builder = builder.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.Email != "" {
		builder = builder.Where(squirrel.Eq{"email": filter.Email})
	}
	if filter.Username != "" {
		builder = builder.Where(squirrel.Eq{"username": filter.Username})
	}

	query, args, err := builder.Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	var u model.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, email, username, passwordHash string) (*model.User, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Insert("users").
		Columns("email", "username", "password_hash").
		Values(email, username, passwordHash).
		Suffix("RETURNING id, email, username, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}

	var u model.User
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&u.ID, &u.Email, &u.Username, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateUser dynamically updates fields of user using model.User object
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, where model.User, update model.User) error {
	if r.db == nil {
		return errors.New("db not connected")
	}

	builder := squtil.PG.Update("users")
	hasUpdates := false

	if update.Email != "" {
		builder = builder.Set("email", update.Email)
		hasUpdates = true
	}
	if update.Username != "" {
		builder = builder.Set("username", update.Username)
		hasUpdates = true
	}
	if update.PasswordHash != "" {
		builder = builder.Set("password_hash", update.PasswordHash)
		hasUpdates = true
	}

	if !hasUpdates {
		return nil
	}

	builder = builder.Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP"))

	// WHERE clause from where model.User
	if where.ID != "" {
		builder = builder.Where(squirrel.Eq{"id": where.ID})
	}
	if where.Email != "" {
		builder = builder.Where(squirrel.Eq{"email": where.Email})
	}
	if where.Username != "" {
		builder = builder.Where(squirrel.Eq{"username": where.Username})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}

// FindUsers fetches paginated users with dynamic filtering using pagination.Fetch
func (r *PostgresUserRepository) FindUsers(ctx context.Context, filter model.User, p pagination.Request) ([]*model.User, *pagination.Meta, error) {
	if r.db == nil {
		return nil, nil, errors.New("db not connected")
	}

	builder := squtil.PG.
		Select("id", "email", "username", "created_at", "updated_at").
		From("users")

	if filter.ID != "" {
		builder = builder.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.Email != "" {
		builder = builder.Where(squirrel.Eq{"email": filter.Email})
	}
	if filter.Username != "" {
		builder = builder.Where(squirrel.Eq{"username": filter.Username})
	}

	allowedSort := map[string]bool{"created_at": true, "email": true, "username": true}

	return pagination.Fetch(
		ctx,
		r.db,
		builder,
		p,
		allowedSort,
		[]string{"created_at DESC"},
		func(rows *sql.Rows) (*model.User, error) {
			var u model.User
			err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.CreatedAt, &u.UpdatedAt)
			return &u, err
		},
	)
}

