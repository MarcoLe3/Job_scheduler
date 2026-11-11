package queue_test

import (
	"context"
	"testing"

	"scheduler/cmd/main/queue"
)

func BenchmarkBaseline(b *testing.B) {
	ch := make(chan int, 1024)
	go func() {
		for range ch {
		}
	}()
	for b.Loop() {
		ch <- 67
	}
}

func BenchmarkQueue(b *testing.B) {
	q := queue.NewBlockingQueue[int](1024)
	ctx := context.Background()
	go func() {
		for {
			if _, err := q.Dequeue(ctx); err != nil {
				return
			}
		}
	}()
	for b.Loop() {
		q.Enqueue(ctx, 67)
	}
}
