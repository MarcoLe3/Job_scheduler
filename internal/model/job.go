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
	ID             string
	IdempotencyKey string
	Type           string
	Payload        json.RawMessage
	Status         Status
	Priority       int
	Attempts       int
	MaxAttempts    int
	RunAt          time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
