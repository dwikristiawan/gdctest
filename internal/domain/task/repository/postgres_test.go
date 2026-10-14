package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/repository"
	"gdctest/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresTaskRepository_CreateTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresTaskRepository(db)
	ctx := context.Background()

	team := &model.Team{
		Name:        "Backend Squad",
		Description: "Handles API and DB",
		OwnerID:     "usr-1",
	}

	rows := sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "created_at", "updated_at"}).
		AddRow("team-1", team.Name, team.Description, team.OwnerID, time.Now(), time.Now())

	mock.ExpectQuery(`INSERT INTO teams \(name,description,owner_id\) VALUES \(\$1,\$2,\$3\) RETURNING id, name, description, owner_id, created_at, updated_at`).
		WithArgs(team.Name, team.Description, team.OwnerID).
		WillReturnRows(rows)

	res, err := repo.CreateTeam(ctx, nil, team)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "team-1", res.ID)
	assert.Equal(t, "Backend Squad", res.Name)
}

func TestPostgresTaskRepository_FindTeamByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresTaskRepository(db)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "created_at", "updated_at"}).
			AddRow("team-1", "Backend Squad", "Desc", "usr-1", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, owner_id, created_at, updated_at FROM teams WHERE id = $1`)).
			WithArgs("team-1").
			WillReturnRows(rows)

		res, err := repo.FindTeamByID(ctx, "team-1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "team-1", res.ID)
	})

	t.Run("not found returns nil", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, owner_id, created_at, updated_at FROM teams WHERE id = $1`)).
			WithArgs("team-ghost").
			WillReturnError(sql.ErrNoRows)

		res, err := repo.FindTeamByID(ctx, "team-ghost")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestPostgresTaskRepository_TeamMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresTaskRepository(db)
	ctx := context.Background()

	t.Run("add member upsert", func(t *testing.T) {
		member := &model.TeamMember{
			TeamID:   "team-1",
			UserID:   "usr-2",
			RoleType: model.RoleTeamMember,
		}

		mock.ExpectExec(`INSERT INTO team_members`).
			WithArgs(member.TeamID, member.UserID, member.RoleType).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AddTeamMember(ctx, nil, member)
		assert.NoError(t, err)
	})

	t.Run("soft delete member", func(t *testing.T) {
		mock.ExpectExec(`UPDATE team_members SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE team_id = \$1 AND user_id = \$2 AND deleted_at IS NULL`).
			WithArgs("team-1", "usr-2").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SoftDeleteTeamMember(ctx, nil, "team-1", "usr-2")
		assert.NoError(t, err)
	})
}

func TestPostgresTaskRepository_TransactionAndLogs(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresTaskRepository(db)
	ctx := context.Background()

	t.Run("transaction commits on success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO task_collaborators`).
			WithArgs("task-1", "usr-2", model.RoleAssignee, "usr-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO task_logs`).
			WithArgs("task-1", nil, "usr-2", "usr-1", "ASSIGN").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		col := &model.TaskCollaborator{
			TaskID:    "task-1",
			UserID:    "usr-2",
			RoleType:  model.RoleAssignee,
			UpdatedBy: func() *string { s := "usr-1"; return &s }(),
		}
		log := &model.TaskLog{
			TaskID:        "task-1",
			NewAssigneeID: "usr-2",
			ChangedBy:     "usr-1",
			Action:        "ASSIGN",
		}

		err := repo.Transaction(ctx, func(tx *sql.Tx) error {
			if err := repo.AssignTaskCollaborator(ctx, tx, col); err != nil {
				return err
			}
			return repo.CreateTaskLog(ctx, tx, log)
		})
		assert.NoError(t, err)
	})

	t.Run("transaction rollbacks on error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO task_collaborators`).
			WithArgs("task-1", "usr-2", model.RoleAssignee, "usr-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO task_logs`).
			WithArgs("task-1", nil, "usr-2", "usr-1", "ASSIGN").
			WillReturnError(errors.New("db error on insert log"))
		mock.ExpectRollback()

		col := &model.TaskCollaborator{
			TaskID:    "task-1",
			UserID:    "usr-2",
			RoleType:  model.RoleAssignee,
			UpdatedBy: func() *string { s := "usr-1"; return &s }(),
		}
		log := &model.TaskLog{
			TaskID:        "task-1",
			NewAssigneeID: "usr-2",
			ChangedBy:     "usr-1",
			Action:        "ASSIGN",
		}

		err := repo.Transaction(ctx, func(tx *sql.Tx) error {
			if err := repo.AssignTaskCollaborator(ctx, tx, col); err != nil {
				return err
			}
			return repo.CreateTaskLog(ctx, tx, log)
		})
		assert.Error(t, err)
		assert.Equal(t, "db error on insert log", err.Error())
	})
}

func TestPostgresTaskRepository_TasksAndPagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	repo := repository.NewPostgresTaskRepository(db)
	ctx := context.Background()

	t.Run("find tasks with pagination", func(t *testing.T) {
		status := model.StatusTodo
		filter := repository.TaskFilter{
			TeamID: "team-1",
			Status: &status,
		}
		p := pagination.Request{
			Page:  1,
			Limit: 10,
			Order: "created_at:desc",
		}

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(SELECT t.id, t.team_id, t.title, t.description, t.status, t.owner_id, t.updated_by, t.created_at, t.updated_at FROM tasks t WHERE t.team_id = \$1 AND t.status = \$2\) AS count_subquery`).
			WithArgs("team-1", status).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`SELECT t.id, t.team_id, t.title, t.description, t.status, t.owner_id, t.updated_by, t.created_at, t.updated_at FROM tasks t WHERE t.team_id = \$1 AND t.status = \$2 ORDER BY created_at DESC, id DESC LIMIT 10 OFFSET 0`).
			WithArgs("team-1", status).
			WillReturnRows(sqlmock.NewRows([]string{"id", "team_id", "title", "description", "status", "owner_id", "updated_by", "created_at", "updated_at"}).
				AddRow("task-1", "team-1", "Setup DB", "Setup postgres schema", status, "usr-1", nil, time.Now(), time.Now()))

		tasks, meta, err := repo.FindTasks(ctx, filter, p)
		assert.NoError(t, err)
		assert.NotNil(t, meta)
		assert.Equal(t, int64(1), meta.TotalRows)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "Setup DB", tasks[0].Title)
	})
}
