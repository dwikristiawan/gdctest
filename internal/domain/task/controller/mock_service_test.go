package controller_test

import (
	"context"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/repository"
	"gdctest/internal/domain/task/service"
	"gdctest/pkg/pagination"

	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTeam(ctx context.Context, ownerID string, req service.CreateTeamRequest) (*model.Team, error) {
	args := m.Called(ctx, ownerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Team), args.Error(1)
}

func (m *MockTaskService) GetTeams(ctx context.Context, userID string) ([]*model.Team, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Team), args.Error(1)
}

func (m *MockTaskService) DeleteTeam(ctx context.Context, teamID, userID string) error {
	args := m.Called(ctx, teamID, userID)
	return args.Error(0)
}

func (m *MockTaskService) AddMember(ctx context.Context, teamID, currentUserID string, req service.AddMemberRequest) error {
	args := m.Called(ctx, teamID, currentUserID, req)
	return args.Error(0)
}

func (m *MockTaskService) GetTeamMembers(ctx context.Context, teamID, currentUserID string) ([]*model.TeamMember, error) {
	args := m.Called(ctx, teamID, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.TeamMember), args.Error(1)
}

func (m *MockTaskService) RemoveMember(ctx context.Context, teamID, currentUserID, targetUserID string) error {
	args := m.Called(ctx, teamID, currentUserID, targetUserID)
	return args.Error(0)
}

func (m *MockTaskService) CreateTask(ctx context.Context, currentUserID string, req service.CreateTaskRequest) (*model.Task, error) {
	args := m.Called(ctx, currentUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskService) GetTasks(ctx context.Context, currentUserID string, filter repository.TaskFilter, p pagination.Request) ([]*model.Task, *pagination.Meta, error) {
	args := m.Called(ctx, currentUserID, filter, p)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var meta *pagination.Meta
	if args.Get(1) != nil {
		meta = args.Get(1).(*pagination.Meta)
	}
	return args.Get(0).([]*model.Task), meta, args.Error(2)
}

func (m *MockTaskService) GetTaskDetail(ctx context.Context, taskID, currentUserID string) (*model.Task, error) {
	args := m.Called(ctx, taskID, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, currentUserID string, req service.UpdateTaskRequest) (*model.Task, error) {
	args := m.Called(ctx, currentUserID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, taskID, currentUserID string) error {
	args := m.Called(ctx, taskID, currentUserID)
	return args.Error(0)
}

func (m *MockTaskService) AssignTask(ctx context.Context, currentUserID string, req service.AssignTaskRequest) error {
	args := m.Called(ctx, currentUserID, req)
	return args.Error(0)
}
