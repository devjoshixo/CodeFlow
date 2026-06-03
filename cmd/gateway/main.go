package main

import (
	"codeflow/internal/execution"
	executionPostgres "codeflow/internal/execution/postgres"
	"codeflow/internal/gateway"
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// authPostgres "codeflow/internal/auth/postgres"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	l := logger.Get()
	cfg, err := config.Load()
	if err != nil {
		l.Error("error in fetchng config", "loaded", false, "error", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		l.Error("failed to connect to db", "err", err.Error())
	}

	if err := pool.Ping(ctx); err != nil {
		l.Error("couldn't ping database")
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	execRepo := executionPostgres.NewExecutionRepo(pool)
	execSvc := execution.NewExecutionService(execRepo, redisClient)

	hub := &gateway.Hub{
		Connections: make(map[string][]*websocket.Conn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ready", readinessHandler(pool, redisClient))
	mux.HandleFunc("GET /ws/executions/{id}", hub.HandleWebSocket(cfg.JWTSecret, execSvc, redisClient))

	server := &http.Server{
		Addr:    ":" + cfg.Gateway_Port,
		Handler: mux,
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		l.Info("shutdown signal received")
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			l.Error("shutdown error", "error", err)
		}
	}()

	l.Info("gateway starting", "port", cfg.Gateway_Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Error("server error", "error", err)
	}
	l.Info("api server stopped gracefully")

}

func readinessHandler(pool *pgxpool.Pool, redis *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("db unavailable\n"))
			return
		}

		if err := redis.Ping(ctx).Err(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("redis unavaiable\n"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready\n"))
	}
}
