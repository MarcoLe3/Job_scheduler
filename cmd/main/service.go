package main

import (
	"context"
	"log"
	"time"

	"scheduler/internal/cache"
	"scheduler/internal/model"
	"scheduler/internal/repo"
)

type Service struct {
	queue *BlockingQueue[*model.Job]
	repo  repo.Repository
	cache *cache.Cache
}

func NewService(r repo.Repository, c *cache.Cache, q *BlockingQueue[*model.Job]) *Service {
	return &Service{
		queue: q,
		repo:  r,
		cache: c,
	}
}

func (s *Service) Submit(ctx context.Context, req model.JobRequest) (*model.Job, error) {
	job := model.NewJob(req)
	if err := s.queue.Offer(job, time.Second*30); err != nil {
		return nil, err
	}

	err := s.repo.Insert(ctx, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Service) Get(ctx context.Context, id string) (*model.Job, error) {
	var job model.Job

	found, err := s.cache.Get(ctx, id, &job)
	if err == nil && found {
		return &job, nil
	}

	j, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, id, j); err != nil {
		log.Printf("cache set failed: %v", err)
	}

	return j, nil
}
