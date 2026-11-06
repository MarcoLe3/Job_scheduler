package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTimeout = errors.New("queue timed out")
	ErrClosed  = errors.New("queue is closed")
)

type BlockingQueue[T any] struct {
	items    []T
	capacity int
	closed   bool
	lock     sync.Mutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
}

func NewBlockingQueue[T any](capacity int) *BlockingQueue[T] {
	q := &BlockingQueue[T]{
		items:    make([]T, 0, capacity),
		capacity: capacity,
	}
	q.notEmpty = sync.NewCond(&q.lock)
	q.notFull = sync.NewCond(&q.lock)
	return q
}

func (q *BlockingQueue[T]) Enqueue(ctx context.Context, item T) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.closed {
		return ErrClosed
	}

	for len(q.items) == q.capacity {
		if err := ctx.Err(); err != nil {
			return err
		}
		q.notFull.Wait()
	}

	q.items = append(q.items, item)
	q.notFull.Signal()
	return nil
}

func (q *BlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	stop := context.AfterFunc(ctx, q.wakeAll)
	defer stop()

	q.lock.Lock()
	defer q.lock.Unlock()

	var zero T
	for len(q.items) == 0 {
		if q.closed {
			return zero, ErrClosed
		}
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		q.notEmpty.Wait()
	}

	item := q.items[0]
	q.items = q.items[1:]
	q.notEmpty.Signal()
	return item, nil
}

func (q *BlockingQueue[T]) Snapshot() []T {
	q.lock.Lock()
	defer q.lock.Unlock()

	out := make([]T, len(q.items))
	copy(out, q.items)
	return out
}

func (q *BlockingQueue[T]) Offer(item T, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := q.Enqueue(ctx, item)
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}

	return err
}

func (q *BlockingQueue[T]) wakeAll() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
}

func (q *BlockingQueue[T]) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.closed = true
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
}
