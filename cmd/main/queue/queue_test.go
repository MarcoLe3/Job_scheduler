package queue_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"scheduler/cmd/main/queue"
)

func TestDequeBlockingUntilEnqueue(t *testing.T) {
	q := queue.NewBlockingQueue[int](2)
	got := make(chan int, 1)

	go func() {
		v, _ := q.Dequeue(context.Background())
		got <- v
	}()

	time.Sleep(time.Microsecond * 50)
	q.Enqueue(context.Background(), 67)

	select {
	case v := <-got:
		if v != 67 {
			t.Fatalf("got %d, want 67", v)
		}
	case <-time.After(time.Second):
		t.Fatal("Dequeue never woke up")
	}
}

func TestEnqueueBlockingUntilDequeue(t *testing.T) {
	q := queue.NewBlockingQueue[int](1)

	if err := q.Enqueue(context.Background(), 67); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- q.Enqueue(context.Background(), 2)
	}()

	select {
	case <-done:
		t.Fatal("Enqueue return while queue is full")
	case <-time.After(time.Millisecond * 50):

	}

	if _, err := q.Dequeue(context.Background()); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Enqueue never woke up")
	}
}

func TestQueueConcurrent(t *testing.T) {
	q := queue.NewBlockingQueue[int](8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const producers, perProducer = 8, 500
	var wg sync.WaitGroup

	for p := 0; p < producers; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perProducer; i++ {
				q.Enqueue(ctx, 67)
			}
		}()
	}

	const consumers = 8

	var mu sync.Mutex
	got := 0
	var cwg sync.WaitGroup

	for c := 0; c < consumers; c++ {
		cwg.Add(1)
		go func() {
			defer cwg.Done()
			for {
				if _, err := q.Dequeue(ctx); err != nil {
					return
				}
				mu.Lock()
				got++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	for len(q.Snapshot()) > 0 {
		time.Sleep(time.Millisecond)
	}

	cancel()
	cwg.Wait()

	if got != producers*perProducer {
		t.Fatalf("got %d item, want %d", got, producers*perProducer)
	}
}
