package controller_test

import (
	"net/http"
	"testing"

	"gdctest/internal/domain/task/controller"
	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskController_CreateTeam(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	body := []byte(`{"name":"Backend Team","description":"Core engineering"}`)
	c, rec := setupEchoTest(http.MethodPost, "/teams", body)
	c.Set("user_id", "user-123")

	mockTeam := &model.Team{
		ID:          "team-1",
		Name:        "Backend Team",
		Description: "Core engineering",
		OwnerID:     "user-123",
	}

	mockSvc.On("CreateTeam", mock.Anything, "user-123", service.CreateTeamRequest{
		Name:        "Backend Team",
		Description: "Core engineering",
	}).Return(mockTeam, nil)

	err := ctrl.CreateTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Backend Team")
}

func TestTaskController_GetTeams(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	c, rec := setupEchoTest(http.MethodGet, "/teams", nil)
	c.Set("user_id", "user-123")

	mockTeams := []*model.Team{
		{ID: "team-1", Name: "Team 1"},
	}
	mockSvc.On("GetTeams", mock.Anything, "user-123").Return(mockTeams, nil)

	err := ctrl.GetTeams(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Team 1")
}

func TestTaskController_DeleteTeam(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	c, rec := setupEchoTest(http.MethodDelete, "/teams/team-1", nil)
	c.Set("user_id", "user-123")
	c.SetParamNames("id")
	c.SetParamValues("team-1")

	mockSvc.On("DeleteTeam", mock.Anything, "team-1", "user-123").Return(nil)

	err := ctrl.DeleteTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "team deleted successfully")
}

func TestTaskController_AddMember(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	body := []byte(`{"user_id":"user-456","role_type":1}`)
	c, rec := setupEchoTest(http.MethodPost, "/teams/team-1/members", body)
	c.Set("user_id", "user-123")
	c.SetParamNames("id")
	c.SetParamValues("team-1")

	mockSvc.On("AddMember", mock.Anything, "team-1", "user-123", service.AddMemberRequest{
		UserID:   "user-456",
		RoleType: 1,
	}).Return(nil)

	err := ctrl.AddMember(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "member added successfully")
}

func TestTaskController_GetTeamMembers(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	c, rec := setupEchoTest(http.MethodGet, "/teams/team-1/members", nil)
	c.Set("user_id", "user-123")
	c.SetParamNames("id")
	c.SetParamValues("team-1")

	mockMembers := []*model.TeamMember{
		{TeamID: "team-1", UserID: "user-456", RoleType: 1},
	}
	mockSvc.On("GetTeamMembers", mock.Anything, "team-1", "user-123").Return(mockMembers, nil)

	err := ctrl.GetTeamMembers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "user-456")
}

func TestTaskController_RemoveMember(t *testing.T) {
	mockSvc := new(MockTaskService)
	ctrl := controller.NewTaskController(mockSvc)

	c, rec := setupEchoTest(http.MethodDelete, "/teams/team-1/members/user-456", nil)
	c.Set("user_id", "user-123")
	c.SetParamNames("id", "user_id")
	c.SetParamValues("team-1", "user-456")

	mockSvc.On("RemoveMember", mock.Anything, "team-1", "user-123", "user-456").Return(nil)

	err := ctrl.RemoveMember(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "member removed successfully")
}
