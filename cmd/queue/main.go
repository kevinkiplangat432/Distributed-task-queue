package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kevinkiplangat432/distributed-task-queue/internal/queue"
	"github.com/kevinkiplangat432/distributed-task-queue/internal/task"
	"github.com/kevinkiplangat432/distributed-task-queue/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	q := queue.New()

	handler := func(ctx context.Context, t *task.Task) error {
		// placeholder handler, real dispatch by t.Type comes later
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	pool := worker.NewPool(q, 3, handler)
	pool.Start(ctx)

	// seed a few demo tasks
	for i := 0; i < 5; i++ {
		_ = q.Push(task.New("demo", []byte("payload")))
	}

	<-ctx.Done()
	log.Println("shutdown signal received, draining workers")

	q.Close()
	pool.Wait()
	log.Println("all workers exited cleanly")
}