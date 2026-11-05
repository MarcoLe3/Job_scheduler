package repo

import (
	"context"

	"scheduler/internal/model"
)

type Repository interface {
	Insert(ctx context.Context, j *model.Job) error
	Get(ctx context.Context, id string) (*model.Job, error)
	MarkDone(ctx context.Context, id string) error
	MarkDead(ctx context.Context, id string) error
	MarkRunning(ctx context.Context, id string) error
}