package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/kevinkiplangat432/distributed-task-queue/internal/task"
)

var ErrClosed = errors.New("queue: closed")

// Queue is a thread-safe FIFO in-memory task queue.
// Redis-backed persistence replaces this in Phase 2.
type Queue struct {
	mu     sync.Mutex
	items  []*task.Task
	notify chan struct{}
	closed bool
}

func New() *Queue {
	return &Queue{
		notify: make(chan struct{}, 1),
	}
}

// Push adds a task to the back of the queue.
func (q *Queue) Push(t *task.Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrClosed
	}
	q.items = append(q.items, t)

	select {
	case q.notify <- struct{}{}:
	default:
	}
	return nil
}

// Pop blocks until a task is available, the context is cancelled, or the queue closes.
func (q *Queue) Pop(ctx context.Context) (*task.Task, error) {
	for {
		q.mu.Lock()
		if len(q.items) > 0 {
			t := q.items[0]
			q.items = q.items[1:]
			q.mu.Unlock()
			return t, nil
		}
		closed := q.closed
		q.mu.Unlock()

		if closed {
			return nil, ErrClosed
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-q.notify:
		}
	}
}

// Len returns the current number of pending tasks.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Close prevents further pushes and wakes any blocked poppers.
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	close(q.notify)
}