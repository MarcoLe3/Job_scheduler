package pool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"scheduler/cmd/main/queue"
	"scheduler/cmd/main/request"
	"scheduler/internal/model"
	"scheduler/internal/repo"
)

type WorkerPool struct {
	queue    *queue.BlockingQueue[*model.Job]
	handlers map[string]request.Request
	repo     repo.Repository
	wg       sync.WaitGroup
}

func NewWorkerPool(q *queue.BlockingQueue[*model.Job], r repo.Repository) *WorkerPool {
	return &WorkerPool{
		queue:    q,
		handlers: make(map[string]request.Request),
		repo:     r,
	}
}

func (p *WorkerPool) AddRequest(jobType string, r request.Request) {
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
	log.Print("worker started")
	for {
		job, err := p.queue.Dequeue(ctx)
		if err != nil {
			log.Print("worker excited")
			return
		}
		p.execute(ctx, job)
	}
}

func (p *WorkerPool) Stop() {
	p.wg.Wait()
}

func (p *WorkerPool) execute(ctx context.Context, job *model.Job) {
	dbCtx := context.WithoutCancel(ctx)

	if err := p.repo.MarkRunning(dbCtx, job.ID, time.Now().UTC()); err != nil {
		log.Printf("job %s: Marking Running failed: %v", job.ID, err)
		return
	}

	job.Attempts++

	jobCtx, cancel := context.WithTimeout(dbCtx, time.Second*30)
	defer cancel()

	err := p.safeRun(jobCtx, job)
	switch {
	case err == nil:
		if e := p.repo.MarkDone(dbCtx, job.ID, time.Now().UTC()); e != nil {
			log.Printf("job %s: Mark Done failed: %v", job.ID, e)
		}
	case job.Attempts < job.MaxAttempts:
		delay := time.Duration(1<<job.Attempts) * time.Second
		log.Printf("job %s failed (attempt %d), retrying in %v: %v", job.ID, job.Attempts, delay, err)
		if e := p.repo.Requeue(dbCtx, job.ID, err.Error(), delay, time.Now().UTC()); e != nil {
			log.Printf("job %s: Requeue failed: %v", job.ID, e)
		}
		time.AfterFunc(delay, func() {
			p.queue.Enqueue(dbCtx, job)
		})
	default:
		log.Printf("job %s dead after %d attempts: %v", job.ID, job.Attempts, err)
		p.repo.MarkDead(dbCtx, job.ID, err.Error(), time.Now().UTC())
	}
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

func (p *WorkerPool) safeRun(ctx context.Context, job *model.Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return p.run(ctx, job)
}
