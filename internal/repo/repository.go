package repo

import (
	"context"
	"time"

	"scheduler/internal/model"
)

type Repository interface {
	Insert(ctx context.Context, j *model.Job) error
	Get(ctx context.Context, id string) (*model.Job, error)
	MarkDone(ctx context.Context, id string, updTime time.Time) error
	MarkDead(ctx context.Context, id string, errMsg string, updTime time.Time) error
	MarkRunning(ctx context.Context, id string, updTime time.Time) error
	MarkFailed(ctx context.Context, id string, errMsg string, updTime time.Time) error
	Requeue(ctx context.Context, id string, errMsg string, delay time.Duration, updTime time.Time) error
}