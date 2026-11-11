package queue

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
	head     int
	count    int
	closed   bool
	lock     sync.Mutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
}

func NewBlockingQueue[T any](capacity int) *BlockingQueue[T] {
	if capacity < 1 {
		capacity = 1
	}

	q := &BlockingQueue[T]{
		items:    make([]T, capacity),
		capacity: capacity,
	}
	q.notEmpty = sync.NewCond(&q.lock)
	q.notFull = sync.NewCond(&q.lock)
	return q
}

func (q *BlockingQueue[T]) Enqueue(ctx context.Context, item T) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	var stop func() bool

	for !q.closed && q.count == q.capacity {
		if err := ctx.Err(); err != nil {
			return err
		}

		if stop == nil {
			stop := context.AfterFunc(ctx, q.wakeAll)
			defer stop()
		}
		q.notFull.Wait()
	}

	if q.closed {
		return ErrClosed
	}

	q.items[(q.head+q.count)%q.capacity] = item
	q.count++
	q.notEmpty.Signal()
	return nil
}

func (q *BlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	var stop func() bool

	var zero T
	for q.count == 0 {
		if q.closed {
			return zero, ErrClosed
		}

		if err := ctx.Err(); err != nil {
			return zero, err
		}

		if stop == nil {
			stop := context.AfterFunc(ctx, q.wakeAll)
			defer stop()
		}

		q.notEmpty.Wait()
	}

	item := q.items[q.head]
	q.items[q.head] = zero
	q.head = (q.head + 1) % q.capacity
	q.count--
	q.notFull.Signal()
	return item, nil
}

func (q *BlockingQueue[T]) Snapshot() []T {
	q.lock.Lock()
	defer q.lock.Unlock()

	out := make([]T, q.count)
	for i := 0; i < q.count; i++ {
		out[i] = q.items[(q.head+i)%q.capacity]
	}
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
