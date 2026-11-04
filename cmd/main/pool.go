package main

import (
	"context"
	"sync"

	"scheduler/internal/model"

	"honnef.co/go/tools/analysis/report"
)

type WorkerPool struct {
	repo 	JobRepository
	wg		sync.WaitGroup
}

func NewWorkerPool(numWorkers int, ctx context.Context) *WorkerPool{
	return &WorkerPool{
		
	}
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
		job, err := p.queue.
	}
}

func (p *WorkerPool) Stop() {
	p.wg.Wait()
}

func (p *WorkerPool) execute(ctx context.Context, job *model.Job) error {
	return p.repo.ExecuteJob(ctx, job)
}