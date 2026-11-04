package main

import (
	"context"

	"scheduler/internal/cache"
	"scheduler/internal/db"
	"scheduler/internal/model"
)

type Service struct {
	queue      *BlockingQueue[*model.Job]
	db         db.DB
	cache      *cache.Cache
	workerPool *WorkerPool
}

func (s *Service) Submit(ctx context.Context, req model.JobRequest) (*model.Job, error) {
	job := model.NewJob(req)
	if err := s.queue.Enqueue(job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Service) Get(ctx context.Context, id string) (*model.Job, error) {
	var job model.Job

	if found, err := s.cache.Get(ctx, id, &job); err != nil {
		return nil, err
	} else if !found {
		return nil, cache.ErrNotFound
	}

	return &job, nil
}
