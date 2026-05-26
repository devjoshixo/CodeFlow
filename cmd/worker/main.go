package main

import (
	"codeflow/internal/execution"
	executionPostgres "codeflow/internal/execution/postgres"
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"codeflow/internal/sandbox"
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger.Init(cfg.Env == "production")
	l := logger.Get()
	l.Info("worker initalized", "env", cfg.Env, "port", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	if err := db.Ping(ctx); err != nil {
		l.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	if err := redisClient.Ping(ctx).Err(); err != nil {
		l.Error("redis ping failed", "error", err)
		os.Exit(1)
	}

	executionRepo := executionPostgres.NewExecutionRepo(db)
	executionService := execution.NewExecutionService(executionRepo, redisClient)

	workerCount := 5
	for i := 0; i < workerCount; i++ {
		go runWorker(ctx, redisClient, executionService, l)
	}

	l.Info("worker started", "count", workerCount)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	l.Info("shutdown signal received, stopping workers")
	cancel()

}

func runWorker(ctx context.Context, redis *redis.Client, svc *execution.ExecutionService, l *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := redis.BRPop(ctx, 0, "executions:queue").Result()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			l.Error("failed to pop from queue", "error", err)
			continue
		}

		executionID := result[1]

		l.Info("job received", "execution_id", executionID)

		exec, err := svc.GetByIDInternal(ctx, executionID)
		if err != nil || exec == nil {
			l.Error("failed to fetch execution", "id", executionID, "error", err)
			continue
		}

		if err := svc.UpdateExecution(ctx, executionID, execution.StatusRunning, "", "", 0, 0); err != nil {
			l.Error("failed to mark running ", "id", executionID, "error", err)
		}

		output, error, exitCode, durationMs, err := sandbox.DockerExecutor(ctx, exec)
		if err != nil {
			l.Error("error in docker execution", "error", err)
			continue
		}

		l.Info("would execute", "language", exec.Language, "id", exec.ID)

		if err := svc.UpdateExecution(ctx, executionID, execution.StatusCompleted, output, error, exitCode, durationMs); err != nil {
			l.Error("failed to mark completed", "id", executionID, "error", err)
		}
	}
}
