package main

import (
	"sync"
)

type BlockingQueue[T any] struct {
	items    []T
	capacity int
	lock     sync.RWMutex
	cond     *sync.Cond
}

func NewBlockingQueue[T any](capacity int) *BlockingQueue[T] {
	return &BlockingQueue[T]{
		items:    make([]T, 0, capacity),
		capacity: capacity,
		lock:     sync.RWMutex{},
		cond:     sync.NewCond(&sync.RWMutex{}),
	}
}

func (q *BlockingQueue[T]) Enqueue(item T) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.items) == q.capacity {
		q.cond.Wait()
	}

	q.items = append(q.items, item)
	q.cond.Signal()
	return nil
}

func (q *BlockingQueue[T]) Dequeue() (T, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.items) == 0 {
		q.cond.Wait()
	}

	item := q.items[0]
	q.items = q.items[1:]
	q.cond.Signal()
	return item, nil
}
