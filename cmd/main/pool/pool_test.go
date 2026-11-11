package pool_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"scheduler/cmd/main/pool"
	"scheduler/cmd/main/queue"
	"scheduler/cmd/main/request"
	"scheduler/internal/model"
	"scheduler/internal/repo"
)

func TestPoolRunEchoJob(t *testing.T) {
	r := repo.NewDemoRepo()
	q := queue.NewBlockingQueue[*model.Job](10)
	pool := pool.NewWorkerPool(q, r)
	pool.AddRequest("echo", request.EchoRequest)

	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx, 2)
	jobReq := &model.JobRequest{Type: "echo", Payload: json.RawMessage(`"message": "hi"`)}
	job := model.NewJob(*jobReq)
	r.Insert(ctx, job)
	q.Enqueue(ctx, job)

	deadline := time.Now().Add(time.Second * 2)
	for time.Now().Before(deadline) {
		got, _ := r.Get(ctx, job.ID)
		if got.Status == model.StatusDone {
			cancel()
			pool.Stop()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("job was not successful")

}
