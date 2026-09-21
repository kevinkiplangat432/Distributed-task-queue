package worker

import (
	"context"
	"log"
	"sync"

	"github.com/kevinkiplangat432/distributed-task-queue/internal/queue"
	"github.com/kevinkiplangat432/distributed-task-queue/internal/task"
)

// Handler processes a task's payload. Returning an error marks the task failed.
type Handler func(ctx context.Context, t *task.Task) error

// Pool runs a fixed number of workers pulling from a shared queue.
type Pool struct {
	q       *queue.Queue
	handler Handler
	size    int
	wg      sync.WaitGroup
}

func NewPool(q *queue.Queue, size int, handler Handler) *Pool {
	return &Pool{
		q:       q,
		handler: handler,
		size:    size,
	}
}

// Start launches the worker goroutines. It returns immediately.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.run(ctx, i)
	}
}

// Wait blocks until all workers have exited.
func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) run(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		t, err := p.q.Pop(ctx)
		if err != nil {
			// context cancelled or queue closed, shut this worker down
			return
		}

		t.SetStatus(task.StatusRunning)

		if err := p.handler(ctx, t); err != nil {
			t.SetStatus(task.StatusFailed)
			log.Printf("worker %d: task %s failed: %v", id, t.ID, err)
			continue
		}

		t.SetStatus(task.StatusCompleted)
		log.Printf("worker %d: task %s completed", id, t.ID)
	}
}