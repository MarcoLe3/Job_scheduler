package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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
		job, err := p.queue.Dequeue(ctx)
		if err != nil {
			return
		}
		p.execute(ctx, job)
	}
}

func (p *WorkerPool) Stop() {
	p.wg.Wait()
}

func (p *WorkerPool) execute(ctx context.Context, job *model.Job) {
	job.Status = model.StatusRunning
	job.Attempts++

	err := p.repo.MarkRunning(ctx, job.ID)
	if err != nil {
		log.Println("job was not successfully marked running")
	}

	err = p.run(ctx, job)
	if err != nil {
		log.Println("job was not successfully ran")
		job.Status = model.StatusFailed
		job.RunAt = time.Now().UTC()
	}

	if job.Attempts >= job.MaxAttempts {
		job.Status = model.StatusDead
		p.repo.MarkDead(ctx, job.ID)
	}

	p.repo.MarkDone(ctx, job.ID)
	job.UpdatedAt = time.Now().UTC()

}

func (p *WorkerPool) run(ctx context.Context, job *model.Job) error {
	req, ok := p.handlers[job.Type]
	if !ok {
		return fmt.Errorf("No handler for request for type %v", job.Type)
	}

	if err := req(ctx, job); err != nil {
		return err
	}
	return nil
}
