package queue_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"scheduler/cmd/main/pool"
	"scheduler/cmd/main/queue"
	"scheduler/cmd/main/request"
	"scheduler/internal/model"
	"scheduler/internal/repo"
)

func Test_QueueRetryRequest(t *testing.T) {
	jReq := &model.JobRequest{
		Type:    "flaky",
		Payload: json.RawMessage(`{"message": "flaky"}`),
	}
	j := model.NewJob(*jReq)
	q := queue.NewBlockingQueue[*model.Job](1)
	r := repo.NewDemoRepo()
	p := pool.NewWorkerPool(q, r)

	const producers, perProducer = 8, 500
	var wg sync.WaitGroup
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			for p := 0; p < perProducer; p++ {
				q.Enqueue(context.Background(), j)
			}
		}()
	}

	wg.Wait()

	p.AddRequest("flaky", request.FlakyRequest)
	p.Start(context.Background(), 8)

}
