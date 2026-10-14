package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gdctest/internal/domain/task/model"
	"gdctest/internal/domain/task/service"
	"gdctest/pkg/apperror"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func computeTestCanonicalHash(req service.CreateTaskRequest) string {
	normalized := map[string]any{
		"team_id":     req.TeamID,
		"title":       strings.TrimSpace(req.Title),
		"description": req.Description,
		"status":      req.Status,
	}
	b, _ := json.Marshal(normalized)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

func TestTaskService_Idempotency_Sequential(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()
	repo := new(MockTaskRepository)
	svc := service.NewTaskService(repo, rdb)

	team := &model.Team{ID: "team-1", OwnerID: "usr-1"}
	createdTask := &model.Task{
		ID:          "task-123",
		TeamID:      "team-1",
		Title:       "Setup CI/CD",
		Description: "Configure pipeline",
		Status:      model.StatusTodo,
		OwnerID:     "usr-1",
	}

	repo.On("FindTeamByID", ctx, "team-1").Return(team, nil)
	// CreateTask harus hanya dipanggil TEPAT 1 KALI
	repo.On("CreateTask", ctx, mock.Anything, mock.Anything).Return(createdTask, nil).Once()

	req := service.CreateTaskRequest{
		TeamID:         "team-1",
		Title:          "Setup CI/CD",
		Description:    "Configure pipeline",
		Status:         model.StatusTodo,
		IdempotencyKey: "11111111-2222-3333-4444-555555555555",
	}

	// Request 1: Berhasil membuat task baru
	res1, err1 := svc.CreateTask(ctx, "usr-1", req)
	assert.NoError(t, err1)
	assert.NotNil(t, res1)
	assert.Equal(t, "task-123", res1.ID)

	// Request 2 (Key sama, payload sama): Mengembalikan task identik tanpa query DB CreateTask kedua
	res2, err2 := svc.CreateTask(ctx, "usr-1", req)
	assert.NoError(t, err2)
	assert.NotNil(t, res2)
	assert.Equal(t, "task-123", res2.ID)
	assert.Equal(t, res1.Title, res2.Title)

	repo.AssertExpectations(t)
}

func TestTaskService_Idempotency_ConflictDifferentPayload(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()
	repo := new(MockTaskRepository)
	svc := service.NewTaskService(repo, rdb)

	team := &model.Team{ID: "team-1", OwnerID: "usr-1"}
	createdTask := &model.Task{
		ID:      "task-123",
		TeamID:  "team-1",
		Title:   "Original Title",
		OwnerID: "usr-1",
	}

	repo.On("FindTeamByID", ctx, "team-1").Return(team, nil)
	repo.On("CreateTask", ctx, mock.Anything, mock.Anything).Return(createdTask, nil).Once()

	idempotencyKey := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	// Request 1: Title "Original Title"
	req1 := service.CreateTaskRequest{
		TeamID:         "team-1",
		Title:          "Original Title",
		IdempotencyKey: idempotencyKey,
	}
	res1, err1 := svc.CreateTask(ctx, "usr-1", req1)
	assert.NoError(t, err1)
	assert.NotNil(t, res1)

	// Request 2: Key SAMA tapi Title BERBEDA ("Modified Title")
	req2 := service.CreateTaskRequest{
		TeamID:         "team-1",
		Title:          "Modified Title",
		IdempotencyKey: idempotencyKey,
	}
	res2, err2 := svc.CreateTask(ctx, "usr-1", req2)
	assert.Error(t, err2)
	assert.Nil(t, res2)

	var appErr *apperror.Error
	assert.ErrorAs(t, err2, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
	assert.Contains(t, err2.Error(), "idempotency key already used with different request parameters")
}

func TestTaskService_Idempotency_ConcurrentDuplicates(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()

	ctx := context.Background()
	repo := new(MockTaskRepository)
	svc := service.NewTaskService(repo, rdb)

	team := &model.Team{ID: "team-1", OwnerID: "usr-1"}
	repo.On("FindTeamByID", ctx, "team-1").Return(team, nil)

	var dbInsertCount int32
	repo.On("CreateTask", ctx, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&dbInsertCount, 1)
			// Berikan jeda kecil untuk mensimulasikan latency I/O DB
			time.Sleep(50 * time.Millisecond)
		}).
		Return(&model.Task{
			ID:          "task-concurrent-1",
			TeamID:      "team-1",
			Title:       "High Concurrency Task",
			Description: "Race condition test",
			OwnerID:     "usr-1",
		}, nil).Once()

	req := service.CreateTaskRequest{
		TeamID:         "team-1",
		Title:          "High Concurrency Task",
		Description:    "Race condition test",
		Status:         model.StatusTodo,
		IdempotencyKey: "99999999-8888-7777-6666-555555555555",
	}

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]*model.Task, goroutines)
	errorsList := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		idx := i
		go func() {
			defer wg.Done()
			results[idx], errorsList[idx] = svc.CreateTask(ctx, "usr-1", req)
		}()
	}

	wg.Wait()

	// ASSERTION 1: Semua goroutine berhasil tanpa error
	for i := 0; i < goroutines; i++ {
		assert.NoError(t, errorsList[i], "goroutine %d failed", i)
		assert.NotNil(t, results[i], "goroutine %d got nil task", i)
		assert.Equal(t, "task-concurrent-1", results[i].ID)
	}

	// ASSERTION 2: INSERT DB HANYA DIJALANKAN TEPAT 1 KALI
	assert.Equal(t, int32(1), atomic.LoadInt32(&dbInsertCount), "DB insert must be called exactly once")
	repo.AssertExpectations(t)
}
