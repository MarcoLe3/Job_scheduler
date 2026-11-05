package main

import (
	"context"
	"sync"

	"scheduler/internal/model"
	"scheduler/internal/repo"
)

type WorkerPool struct {
	queue    *BlockingQueue[*model.Job]
	handlers map[string]Request
	repo     repo.Repository
	wg       sync.WaitGroup
}

func NewWorkerPool(q *BlockingQueue[*model.Job], r repo.Repository) *WorkerPool {
	return &WorkerPool{
		queue:    q,
		handlers: make(map[string]Request),
		repo:     r,
	}
}

func (p *WorkerPool) AddRequest(jobType string, r Request) {
	p.handlers[jobType] = r
}

func (p *WorkerPool) Start(ctx context.Context, n int) {
	for i := 0; i < n; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *WorkerPool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		job, err := p.queue.Dequeue()
		if err != nil {
			return
		}
		p.execute(ctx, job)
	}
}

func (p *WorkerPool) Stop() {
	p.wg.Wait()
}

func (p *WorkerPool) execute(ctx context.Context, job *model.Job) error {
	jobType := job.Type
	req := p.handlers[jobType]
	if err := req(ctx, job); err != nil {
		return err
	}
	return nil
}
