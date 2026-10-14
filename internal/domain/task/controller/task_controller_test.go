package controller_test

import (
	"net/http"
	"testing"

	"gdctest/internal/domain/task/controller"
	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/repository"
	"gdctest/internal/domain/task/service"
	"gdctest/pkg/pagination"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskController_CreateTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	body := []byte(`{"team_id":"team-1","title":"New Feature","description":"Task desc","status":0}`)
	c, rec := setupEchoTest(http.MethodPost, "/tasks", body)
	c.Set("user_id", "user-123")
	c.Request().Header.Set("Idempotency-Key", "11111111-2222-3333-4444-555555555555")

	mockTask := &model.Task{
		ID:          "task-1",
		TeamID:      "team-1",
		Title:       "New Feature",
		Description: "Task desc",
		OwnerID:     "user-123",
	}

	mockSvc.On("CreateTask", mock.Anything, "user-123", service.CreateTaskRequest{
		TeamID:         "team-1",
		Title:          "New Feature",
		Description:    "Task desc",
		Status:         0,
		IdempotencyKey: "11111111-2222-3333-4444-555555555555",
	}).Return(mockTask, nil)

	err := ctrl.CreateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "New Feature")
}

func TestTaskController_GetTasks(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	c, rec := setupEchoTest(http.MethodGet, "/tasks?team_id=team-1&page=1&limit=10", nil)
	c.Set("user_id", "user-123")

	mockTasks := []*model.Task{
		{ID: "task-1", Title: "Task 1"},
	}
	meta := &pagination.Meta{Page: 1, Limit: 10, TotalRows: 1, TotalPages: 1}

	mockSvc.On("GetTasks", mock.Anything, "user-123", repository.TaskFilter{
		TeamID: "team-1",
	}, pagination.Request{Page: 1, Limit: 10}).Return(mockTasks, meta, nil)

	err := ctrl.GetTasks(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Task 1")
}

func TestTaskController_GetTaskDetail(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	c, rec := setupEchoTest(http.MethodGet, "/tasks/task-1", nil)
	c.Set("user_id", "user-123")
	c.SetParamNames("id")
	c.SetParamValues("task-1")

	mockTask := &model.Task{ID: "task-1", Title: "Task 1"}
	mockSvc.On("GetTaskDetail", mock.Anything, "task-1", "user-123").Return(mockTask, nil)

	err := ctrl.GetTaskDetail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Task 1")
}

func TestTaskController_UpdateTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	body := []byte(`{"title":"Updated Title"}`)
	c, rec := setupEchoTest(http.MethodPut, "/tasks/task-1", body)
	c.Set("user_id", "user-123")
	c.SetParamNames("id")
	c.SetParamValues("task-1")

	title := "Updated Title"
	mockTask := &model.Task{ID: "task-1", Title: title}

	mockSvc.On("UpdateTask", mock.Anything, "user-123", service.UpdateTaskRequest{
		ID:    "task-1",
		Title: &title,
	}).Return(mockTask, nil)

	err := ctrl.UpdateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Updated Title")
}

func TestTaskController_DeleteTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	c, rec := setupEchoTest(http.MethodDelete, "/tasks/task-1", nil)
	c.Set("user_id", "user-123")
	c.SetParamNames("id")
	c.SetParamValues("task-1")

	mockSvc.On("DeleteTask", mock.Anything, "task-1", "user-123").Return(nil)

	err := ctrl.DeleteTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "task deleted successfully")
}

func TestTaskController_AssignTask(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	body := []byte(`{"user_id":"user-456","role_type":0}`)
	c, rec := setupEchoTest(http.MethodPost, "/tasks/task-1/assign", body)
	c.Set("user_id", "user-123")
	c.SetParamNames("id")
	c.SetParamValues("task-1")

	mockSvc.On("AssignTask", mock.Anything, "user-123", service.AssignTaskRequest{
		TaskID:   "task-1",
		UserID:   "user-456",
		RoleType: 0,
	}).Return(nil)

	err := ctrl.AssignTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "task collaborator assigned successfully")
}
