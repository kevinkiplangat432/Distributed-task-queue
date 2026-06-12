package queue

import (
	"sync"

	"github.com/yourusername/distributed-task-queue/internal/task"
)

type Task = task.Task

type Queue interface {
	Enqueue(Task)
	Dequeue() *Task
	Len() int
}

type inMemoryQueue struct {
	mu    sync.Mutex
	tasks []Task
}

func NewInMemoryQueue() Queue {
	return &inMemoryQueue{}
}

func (q *inMemoryQueue) Enqueue(t Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, t)
}

func (q *inMemoryQueue) Dequeue() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tasks) == 0 {
		return nil
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return &task
}

func (q *inMemoryQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tasks)
}
