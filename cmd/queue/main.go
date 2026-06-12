package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/distributed-task-queue/internal/queue"
	"github.com/yourusername/distributed-task-queue/internal/task"
	"github.com/yourusername/distributed-task-queue/internal/worker"
)

func main() {
	role := flag.String("role", "dispatcher", "service role: dispatcher or worker")
	interval := flag.Duration("interval", 5*time.Second, "worker polling interval")
	flag.Parse()

	switch *role {
	case "dispatcher":
		runDispatcher()
	case "worker":
		runWorker(*interval)
	default:
		log.Fatalf("invalid role: %s; expected dispatcher or worker", *role)
	}
}

func runDispatcher() {
	q := queue.NewInMemoryQueue()

	tasks := []task.Task{
		task.NewTask("task-1", "generate-report", nil),
		task.NewTask("task-2", "send-email", map[string]interface{}{"recipient": "user@example.com"}),
	}

	for _, t := range tasks {
		q.Enqueue(t)
		fmt.Printf("queued task %s (%s)\n", t.ID, t.Type)
	}

	fmt.Println("dispatcher ready; workers can now dequeue tasks")
}

func runWorker(interval time.Duration) {
	w := worker.NewWorker("worker-1")
	fmt.Printf("starting worker %s\n", w.ID)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		task, err := w.FetchTask()
		if err != nil {
			log.Printf("worker fetch failed: %v", err)
			continue
		}

		if task == nil {
			log.Println("no tasks available; waiting")
			continue
		}

		err = w.Execute(task)
		if err != nil {
			log.Printf("task execution failed: %v", err)
			continue
		}

		fmt.Printf("worker %s completed task %s\n", w.ID, task.ID)
	}
}
