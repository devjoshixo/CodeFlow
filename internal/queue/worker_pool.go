package queue

import (
	"context"
	"log/slog"
	"time"
)

type WorkerPool struct {
	queue   *Queue
	workers int
	handler func(context.Context, *Job) error
	logger  *slog.Logger
}

func NewWorkerPool(queue *Queue, workers int, handler func(context.Context, *Job) error, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{queue: queue, workers: workers, handler: handler, logger: logger}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		go wp.runWorker(ctx, i)
	}
}

func (wp *WorkerPool) runWorker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			wp.logger.Info("worker stopped", "id", id)
			return
		default:
		}

		job := wp.queue.Pop()
		wp.logger.Info("job received", "id", job.ID, "priority", job.Priority)

		jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := wp.handler(jobCtx, job)
		cancel()

		if err != nil {
			job.Retries++
			if job.Retries >= 5 {
				wp.logger.Error("job abandoned after 5 retries", "id", job.ID)
			}
			wp.logger.Error("job failed", "id", job.ID, "retries", job.Retries, "error", err)
			continue
		}
		backoff := time.Duration(1<<(job.Retries-1)) * time.Second
		wp.logger.Info("job queued", "id", job.ID, "retries", job.Retries, "backoff", backoff)

		time.Sleep(backoff)
		wp.queue.Push(job)
		continue
	}
}
