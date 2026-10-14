package repository

import (
	"context"
	"database/sql"
	"errors"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/pagination"
	"gdctest/pkg/squtil"

	"github.com/Masterminds/squirrel"
)

type PostgresTaskRepository struct {
	db *sql.DB
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *PostgresTaskRepository) getExecutor(tx *sql.Tx) dbExecutor {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *PostgresTaskRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}
	return r.db.BeginTx(ctx, nil)
}

func (r *PostgresTaskRepository) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func NewPostgresTaskRepository(db *sql.DB) TaskRepositoryInterface {
	return &PostgresTaskRepository{db: db}
}

// ==========================================
// Teams
// ==========================================

func (r *PostgresTaskRepository) CreateTeam(ctx context.Context, tx *sql.Tx, team *model.Team) (*model.Team, error) {
	exec := r.getExecutor(tx)
	if exec == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Insert("teams").
		Columns("name", "description", "owner_id").
		Values(team.Name, team.Description, team.OwnerID).
		Suffix("RETURNING id, name, description, owner_id, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}

	var res model.Team
	err = exec.QueryRowContext(ctx, query, args...).Scan(
		&res.ID, &res.Name, &res.Description, &res.OwnerID, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *PostgresTaskRepository) FindTeamByID(ctx context.Context, teamID string) (*model.Team, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Select("id", "name", "description", "owner_id", "created_at", "updated_at").
		From("teams").
		Where(squirrel.Eq{"id": teamID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var t model.Team
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&t.ID, &t.Name, &t.Description, &t.OwnerID, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTaskRepository) FindUserTeams(ctx context.Context, userID string) ([]*model.Team, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Select("t.id", "t.name", "t.description", "t.owner_id", "t.created_at", "t.updated_at").
		From("teams t").
		Join("team_members tm ON tm.team_id = t.id").
		Where(squirrel.Eq{"tm.user_id": userID}).
		Where(squirrel.Expr("tm.deleted_at IS NULL")).
		OrderBy("t.created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*model.Team
	for rows.Next() {
		var t model.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.OwnerID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, &t)
	}
	return teams, nil
}

func (r *PostgresTaskRepository) DeleteTeam(ctx context.Context, tx *sql.Tx, teamID string) error {
	exec := r.getExecutor(tx)
	if exec == nil {
		return errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Delete("teams").
		Where(squirrel.Eq{"id": teamID}).
		ToSql()
	if err != nil {
		return err
	}

	res, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("team not found")
	}
	return nil
}

// ==========================================
// Team Members
// ==========================================

func (r *PostgresTaskRepository) AddTeamMember(ctx context.Context, tx *sql.Tx, member *model.TeamMember) error {
	exec := r.getExecutor(tx)
	if exec == nil {
		return errors.New("db not connected")
	}

	// Upsert: Jika sudah ada (misal pernah soft delete), re-aktifkan (deleted_at = NULL)
	query := `
		INSERT INTO team_members (team_id, user_id, role_type, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)
		ON CONFLICT (team_id, user_id) DO UPDATE
		SET role_type = EXCLUDED.role_type,
		    updated_at = CURRENT_TIMESTAMP,
		    deleted_at = NULL
	`

	_, err := exec.ExecContext(ctx, query, member.TeamID, member.UserID, member.RoleType)
	return err
}

func (r *PostgresTaskRepository) FindTeamMember(ctx context.Context, teamID, userID string) (*model.TeamMember, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Select("team_id", "user_id", "role_type", "created_at", "updated_at", "deleted_at").
		From("team_members").
		Where(squirrel.Eq{"team_id": teamID, "user_id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var m model.TeamMember
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&m.TeamID, &m.UserID, &m.RoleType, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PostgresTaskRepository) FindActiveTeamMembers(ctx context.Context, teamID string) ([]*model.TeamMember, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Select("tm.team_id", "tm.user_id", "tm.role_type", "tm.created_at", "tm.updated_at", "u.email", "u.username").
		From("team_members tm").
		Join("users u ON u.id = tm.user_id").
		Where(squirrel.Eq{"tm.team_id": teamID}).
		Where(squirrel.Expr("tm.deleted_at IS NULL")).
		OrderBy("tm.created_at ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.TeamMember
	for rows.Next() {
		var m model.TeamMember
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.RoleType, &m.CreatedAt, &m.UpdatedAt, &m.Email, &m.Username); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *PostgresTaskRepository) SoftDeleteTeamMember(ctx context.Context, tx *sql.Tx, teamID, userID string) error {
	exec := r.getExecutor(tx)
	if exec == nil {
		return errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Update("team_members").
		Set("deleted_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Where(squirrel.Eq{"team_id": teamID, "user_id": userID}).
		Where(squirrel.Expr("deleted_at IS NULL")).
		ToSql()
	if err != nil {
		return err
	}

	res, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("member not found or already removed")
	}
	return nil
}

// ==========================================
// Tasks
// ==========================================

func (r *PostgresTaskRepository) CreateTask(ctx context.Context, tx *sql.Tx, task *model.Task) (*model.Task, error) {
	exec := r.getExecutor(tx)
	if exec == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Insert("tasks").
		Columns("team_id", "title", "description", "status", "owner_id").
		Values(task.TeamID, task.Title, task.Description, task.Status, task.OwnerID).
		Suffix("RETURNING id, team_id, title, description, status, owner_id, updated_by, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}

	var res model.Task
	err = exec.QueryRowContext(ctx, query, args...).Scan(
		&res.ID, &res.TeamID, &res.Title, &res.Description, &res.Status,
		&res.OwnerID, &res.UpdatedBy, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *PostgresTaskRepository) FindTaskByID(ctx context.Context, taskID string) (*model.Task, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Select("id", "team_id", "title", "description", "status", "owner_id", "updated_by", "created_at", "updated_at").
		From("tasks").
		Where(squirrel.Eq{"id": taskID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var t model.Task
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status,
		&t.OwnerID, &t.UpdatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PostgresTaskRepository) FindTasks(ctx context.Context, filter TaskFilter, p pagination.Request) ([]*model.Task, *pagination.Meta, error) {
	if r.db == nil {
		return nil, nil, errors.New("db not connected")
	}

	builder := squtil.PG.
		Select("t.id", "t.team_id", "t.title", "t.description", "t.status", "t.owner_id", "t.updated_by", "t.created_at", "t.updated_at").
		From("tasks t")

	if filter.TeamID != "" {
		builder = builder.Where(squirrel.Eq{"t.team_id": filter.TeamID})
	}
	if filter.Status != nil {
		builder = builder.Where(squirrel.Eq{"t.status": *filter.Status})
	}
	if filter.OwnerID != "" {
		builder = builder.Where(squirrel.Eq{"t.owner_id": filter.OwnerID})
	}
	if filter.AssigneeID != "" {
		builder = builder.
			Join("task_collaborators tc ON tc.task_id = t.id").
			Where(squirrel.Eq{"tc.user_id": filter.AssigneeID, "tc.role_type": model.RoleAssignee})
	}
	if filter.Search != "" {
		builder = builder.Where(squirrel.ILike{"t.title": "%" + filter.Search + "%"})
	}

	allowedSort := map[string]bool{"created_at": true, "updated_at": true, "title": true, "status": true}

	return pagination.Fetch(
		ctx,
		r.db,
		builder,
		p,
		allowedSort,
		[]string{"t.created_at DESC"},
		func(rows *sql.Rows) (*model.Task, error) {
			var t model.Task
			err := rows.Scan(
				&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status,
				&t.OwnerID, &t.UpdatedBy, &t.CreatedAt, &t.UpdatedAt,
			)
			return &t, err
		},
	)
}

func (r *PostgresTaskRepository) UpdateTask(ctx context.Context, tx *sql.Tx, task *model.Task) error {
	exec := r.getExecutor(tx)
	if exec == nil {
		return errors.New("db not connected")
	}

	builder := squtil.PG.
		Update("tasks").
		Set("title", task.Title).
		Set("description", task.Description).
		Set("status", task.Status).
		Set("updated_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Where(squirrel.Eq{"id": task.ID})

	if task.UpdatedBy != nil {
		builder = builder.Set("updated_by", *task.UpdatedBy)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	res, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("task not found")
	}
	return nil
}

func (r *PostgresTaskRepository) DeleteTask(ctx context.Context, tx *sql.Tx, taskID string) error {
	exec := r.getExecutor(tx)
	if exec == nil {
		return errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Delete("tasks").
		Where(squirrel.Eq{"id": taskID}).
		ToSql()
	if err != nil {
		return err
	}

	res, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("task not found")
	}
	return nil
}

// ==========================================
// Task Collaborators & Logs
// ==========================================

func (r *PostgresTaskRepository) AssignTaskCollaborator(ctx context.Context, tx *sql.Tx, col *model.TaskCollaborator) error {
	exec := r.getExecutor(tx)
	if exec == nil {
		return errors.New("db not connected")
	}

	query := `
		INSERT INTO task_collaborators (task_id, user_id, role_type, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (task_id, user_id, role_type) DO UPDATE
		SET updated_by = EXCLUDED.updated_by,
		    updated_at = CURRENT_TIMESTAMP
	`
	_, err := exec.ExecContext(ctx, query, col.TaskID, col.UserID, col.RoleType, col.UpdatedBy)
	return err
}

func (r *PostgresTaskRepository) CreateTaskLog(ctx context.Context, tx *sql.Tx, log *model.TaskLog) error {
	exec := r.getExecutor(tx)
	if exec == nil {
		return errors.New("db not connected")
	}

	query := `
		INSERT INTO task_logs (task_id, previous_assignee_id, new_assignee_id, changed_by, action, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`
	_, err := exec.ExecContext(ctx, query, log.TaskID, log.PreviousAssigneeID, log.NewAssigneeID, log.ChangedBy, log.Action)
	return err
}

func (r *PostgresTaskRepository) FindTaskCollaborators(ctx context.Context, taskID string) ([]*model.TaskCollaborator, error) {
	if r.db == nil {
		return nil, errors.New("db not connected")
	}

	query, args, err := squtil.PG.
		Select("tc.task_id", "tc.user_id", "tc.role_type", "tc.updated_by", "tc.created_at", "tc.updated_at", "u.email", "u.username").
		From("task_collaborators tc").
		Join("users u ON u.id = tc.user_id").
		Where(squirrel.Eq{"tc.task_id": taskID}).
		OrderBy("tc.created_at ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []*model.TaskCollaborator
	for rows.Next() {
		var c model.TaskCollaborator
		if err := rows.Scan(&c.TaskID, &c.UserID, &c.RoleType, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt, &c.Email, &c.Username); err != nil {
			return nil, err
		}
		cols = append(cols, &c)
	}
	return cols, nil
}
