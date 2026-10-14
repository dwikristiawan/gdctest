package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gdctest/internal/domain/task/model"
	"gdctest/pkg/apperror"
)

type idempotencyRecord struct {
	Hash string      `json:"hash"`
	Task *model.Task `json:"task"`
}

func computeCanonicalHash(req CreateTaskRequest) string {
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

func (s *TaskService) CreateTask(ctx context.Context, currentUserID string, req CreateTaskRequest) (*model.Task, error) {
	if req.TeamID == "" {
		return nil, apperror.BadRequest("team_id is required")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, apperror.BadRequest("task title is required")
	}

	var incomingHash string
	var redisKey string
	if req.IdempotencyKey != "" && s.rdb != nil {
		incomingHash = computeCanonicalHash(req)
		redisKey = "idempotency:task:" + req.IdempotencyKey

		// 1. Cek cache
		cachedVal, err := s.rdb.Get(ctx, redisKey).Result()
		if err == nil {
			if strings.HasPrefix(cachedVal, "PROCESSING:") {
				for i := 0; i < 30; i++ {
					time.Sleep(100 * time.Millisecond)
					cachedVal, err = s.rdb.Get(ctx, redisKey).Result()
					if err == nil && !strings.HasPrefix(cachedVal, "PROCESSING:") {
						break
					}
				}
			}

			if err == nil && !strings.HasPrefix(cachedVal, "PROCESSING:") {
				var rec idempotencyRecord
				if err := json.Unmarshal([]byte(cachedVal), &rec); err == nil {
					if rec.Hash != incomingHash {
						return nil, apperror.New(http.StatusConflict, "idempotency key already used with different request parameters")
					}
					return rec.Task, nil
				}
			}
		}

		// 2. Concurrency lock dengan SETNX
		locked, err := s.rdb.SetNX(ctx, redisKey, "PROCESSING:"+incomingHash, 24*time.Hour).Result()
		if err == nil && !locked {
			// Kalah race condition: tunggu hingga pemenang selesai dan ambil hasilnya
			for i := 0; i < 30; i++ {
				time.Sleep(100 * time.Millisecond)
				cachedVal, err = s.rdb.Get(ctx, redisKey).Result()
				if err == nil && !strings.HasPrefix(cachedVal, "PROCESSING:") {
					var rec idempotencyRecord
					if err := json.Unmarshal([]byte(cachedVal), &rec); err == nil {
						if rec.Hash != incomingHash {
							return nil, apperror.New(http.StatusConflict, "idempotency key already used with different request parameters")
						}
						return rec.Task, nil
					}
				}
			}
		}
	}

	team, err := s.repo.FindTeamByID(ctx, req.TeamID)
	if err != nil {
		if redisKey != "" && s.rdb != nil {
			_ = s.rdb.Del(ctx, redisKey).Err()
		}
		return nil, apperror.Internal("failed to fetch team", err)
	}
	if team == nil {
		if redisKey != "" && s.rdb != nil {
			_ = s.rdb.Del(ctx, redisKey).Err()
		}
		return nil, apperror.NotFound("team not found")
	}

	// Hanya anggota tim aktif atau owner tim yang boleh membuat task
	if team.OwnerID != currentUserID {
		member, err := s.repo.FindTeamMember(ctx, req.TeamID, currentUserID)
		if err != nil {
			if redisKey != "" && s.rdb != nil {
				_ = s.rdb.Del(ctx, redisKey).Err()
			}
			return nil, apperror.Internal("failed to verify membership", err)
		}
		if member == nil || member.DeletedAt != nil {
			if redisKey != "" && s.rdb != nil {
				_ = s.rdb.Del(ctx, redisKey).Err()
			}
			return nil, apperror.Forbidden("only active team members can create tasks")
		}
	}

	status := req.Status
	if status < model.StatusBacklog || status > model.StatusArchived {
		status = model.StatusBacklog
	}

	task := &model.Task{
		TeamID:      req.TeamID,
		Title:       title,
		Description: req.Description,
		Status:      status,
		OwnerID:     currentUserID,
	}

	var createdTask *model.Task
	err = s.repo.Transaction(ctx, func(tx *sql.Tx) error {
		var txErr error
		createdTask, txErr = s.repo.CreateTask(ctx, tx, task)
		return txErr
	})
	if err != nil {
		if redisKey != "" && s.rdb != nil {
			_ = s.rdb.Del(ctx, redisKey).Err()
		}
		return nil, apperror.Internal("failed to create task", err)
	}

	if redisKey != "" && s.rdb != nil {
		rec := idempotencyRecord{
			Hash: incomingHash,
			Task: createdTask,
		}
		if b, err := json.Marshal(rec); err == nil {
			_ = s.rdb.Set(ctx, redisKey, string(b), 24*time.Hour).Err()
		}
	}

	return createdTask, nil
}
