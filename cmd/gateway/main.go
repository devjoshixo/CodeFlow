package main

import (
	"codeflow/internal/execution"
	executionPostgres "codeflow/internal/execution/postgres"
	"codeflow/internal/gateway"
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"context"
	"fmt"
	"net/http"

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
	mux.HandleFunc("GET /ws/executions/{id}", hub.HandleWebSocket(cfg.JWTSecret, execSvc, redisClient))

	l.Info("gateway starting", "port", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Gateway_Port), mux); err != nil {
		l.Error("server error", "error", err)
	}

}
