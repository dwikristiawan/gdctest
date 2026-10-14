package service_test

import (
	"context"
	"database/sql"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/repository"
	"gdctest/pkg/pagination"

	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *MockTaskRepository) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "Transaction" {
			args := m.Called(ctx, mock.Anything)
			if err := args.Error(0); err != nil {
				return err
			}
			return fn(nil)
		}
	}
	return fn(nil)
}

func (m *MockTaskRepository) CreateTeam(ctx context.Context, tx *sql.Tx, team *model.Team) (*model.Team, error) {
	for _, call := range m.ExpectedCalls {
		if call.Method == "CreateTeam" && len(call.Arguments) == 2 {
			args := m.Called(ctx, team)
			if args.Get(0) == nil {
				return nil, args.Error(1)
			}
			return args.Get(0).(*model.Team), args.Error(1)
		}
	}
	args := m.Called(ctx, tx, team)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}

func (m *MockTaskRepository) FindTeamByID(ctx context.Context, teamID string) (*model.Team, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}

func (m *MockTaskRepository) FindUserTeams(ctx context.Context, userID string) ([]*model.Team, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Team), args.Error(1)
}

func (m *MockTaskRepository) DeleteTeam(ctx context.Context, tx *sql.Tx, teamID string) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "DeleteTeam" && len(call.Arguments) == 2 {
			return m.Called(ctx, teamID).Error(0)
		}
	}
	return m.Called(ctx, tx, teamID).Error(0)
}

func (m *MockTaskRepository) AddTeamMember(ctx context.Context, tx *sql.Tx, member *model.TeamMember) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "AddTeamMember" && len(call.Arguments) == 2 {
			return m.Called(ctx, member).Error(0)
		}
	}
	return m.Called(ctx, tx, member).Error(0)
}

func (m *MockTaskRepository) FindTeamMember(ctx context.Context, teamID, userID string) (*model.TeamMember, error) {
	args := m.Called(ctx, teamID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TeamMember), args.Error(1)
}

func (m *MockTaskRepository) FindActiveTeamMembers(ctx context.Context, teamID string) ([]*model.TeamMember, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.TeamMember), args.Error(1)
}

func (m *MockTaskRepository) SoftDeleteTeamMember(ctx context.Context, tx *sql.Tx, teamID, userID string) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "SoftDeleteTeamMember" && len(call.Arguments) == 3 {
			return m.Called(ctx, teamID, userID).Error(0)
		}
	}
	return m.Called(ctx, tx, teamID, userID).Error(0)
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, tx *sql.Tx, task *model.Task) (*model.Task, error) {
	for _, call := range m.ExpectedCalls {
		if call.Method == "CreateTask" && len(call.Arguments) == 2 {
			args := m.Called(ctx, task)
			if args.Get(0) == nil {
				return nil, args.Error(1)
			}
			return args.Get(0).(*model.Task), args.Error(1)
		}
	}
	args := m.Called(ctx, tx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepository) FindTaskByID(ctx context.Context, taskID string) (*model.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepository) FindTasks(ctx context.Context, filter repository.TaskFilter, p pagination.Request) ([]*model.Task, *pagination.Meta, error) {
	args := m.Called(ctx, filter, p)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var meta *pagination.Meta
	if args.Get(1) != nil {
		meta = args.Get(1).(*pagination.Meta)
	}
	return args.Get(0).([]*model.Task), meta, args.Error(2)
}

func (m *MockTaskRepository) UpdateTask(ctx context.Context, tx *sql.Tx, task *model.Task) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "UpdateTask" && len(call.Arguments) == 2 {
			return m.Called(ctx, task).Error(0)
		}
	}
	return m.Called(ctx, tx, task).Error(0)
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, tx *sql.Tx, taskID string) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "DeleteTask" && len(call.Arguments) == 2 {
			return m.Called(ctx, taskID).Error(0)
		}
	}
	return m.Called(ctx, tx, taskID).Error(0)
}

func (m *MockTaskRepository) AssignTaskCollaborator(ctx context.Context, tx *sql.Tx, col *model.TaskCollaborator) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "AssignTaskCollaborator" && len(call.Arguments) == 2 {
			return m.Called(ctx, col).Error(0)
		}
	}
	return m.Called(ctx, tx, col).Error(0)
}

func (m *MockTaskRepository) CreateTaskLog(ctx context.Context, tx *sql.Tx, log *model.TaskLog) error {
	for _, call := range m.ExpectedCalls {
		if call.Method == "CreateTaskLog" && len(call.Arguments) == 2 {
			return m.Called(ctx, log).Error(0)
		}
	}
	return m.Called(ctx, tx, log).Error(0)
}

func (m *MockTaskRepository) FindTaskCollaborators(ctx context.Context, taskID string) ([]*model.TaskCollaborator, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.TaskCollaborator), args.Error(1)
}
