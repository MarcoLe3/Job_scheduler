package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusRejected  Status = "rejected"
	StatusDead      Status = "dead"
)

type JobRequest struct {
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	Priority       int             `json:"priority"`
	IdempotencyKey string          `json:"idempotency_key"`
	RunAt          string          `json:"run_at"`
}

type Job struct {
	ID             string          `json:"id"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	Status         Status          `json:"status"`
	Priority       int             `json:"priority"`
	Attempts       int             `json:"attempts"`
	MaxAttempts    int             `json:"max_attempts"`
	LastError      string          `json:"last_error,omitempty"`
	RunAt          time.Time       `json:"run_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func NewJob(req JobRequest) *Job {
	now := time.Now().UTC()
	return &Job{
		ID:             uuid.NewString(),
		IdempotencyKey: req.IdempotencyKey,
		Type:           req.Type,
		Payload:        req.Payload,
		Status:         StatusQueued,
		Priority:       req.Priority,
		Attempts:       0,
		MaxAttempts:    3,
		RunAt:          now,
	}
}
