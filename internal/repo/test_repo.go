package repo

import (
	"context"
	"errors"
	"sync"
	"time"

	"scheduler/internal/model"
)

var (
	ErrNotFound = errors.New("Not found")
)

type DemoRepo struct {
	mu   sync.Mutex
	jobs map[string]*model.Job
}

func NewDemoRepo() *DemoRepo {
	return &DemoRepo{
		jobs: make(map[string]*model.Job),
	}
}

func (r *DemoRepo) Insert(ctx context.Context, j *model.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c := *j
	r.jobs[c.ID] = &c
	return nil
}

func (r *DemoRepo) Get(ctx context.Context, id string) (*model.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	j, ok := r.jobs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return j, nil
}

func (r *DemoRepo) update(id string, updTime time.Time, fn func(*model.Job)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	j, ok := r.jobs[id]
	if !ok {
		return ErrNotFound
	}

	fn(j)
	j.UpdatedAt = updTime
	return nil
}

func (r *DemoRepo) MarkDone(ctx context.Context, id string, updTime time.Time) error {
	return r.update(id, updTime, func(j *model.Job) {
		j.Status = model.StatusDone
	})
}

func (r *DemoRepo) MarkDead(ctx context.Context, id string, errMsg string, updTime time.Time) error {
	return r.update(id, updTime, func(j *model.Job) {
		j.Status = model.StatusDead
		j.LastError = errMsg
	})
}

func (r *DemoRepo) MarkRunning(ctx context.Context, id string, updTime time.Time) error {
	return r.update(id, updTime, func(j *model.Job) {
		j.Status = model.StatusRunning
		j.Attempts++
	})
}

func (r *DemoRepo) MarkFailed(ctx context.Context, id string, errMsg string, updTime time.Time) error {
	return r.update(id, updTime, func(j *model.Job) {
		j.Status = model.StatusFailed
		j.LastError = errMsg
	})
}

func (r *DemoRepo) Requeue(ctx context.Context, id string, errMsg string, delay time.Duration, updTime time.Time) error {
	return r.update(id, updTime, func(j *model.Job) {
		j.Status = model.StatusQueued
		j.LastError = errMsg
		j.UpdatedAt = updTime
		j.RunAt = time.Now().UTC().Add(delay)
	})
}
