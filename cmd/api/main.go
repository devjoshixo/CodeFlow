package main

import (
	"codeflow/internal/auth"
	"codeflow/internal/auth/postgres"
	"codeflow/internal/execution"
	executionPostgres "codeflow/internal/execution/postgres"
	"codeflow/internal/metrics"
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"codeflow/internal/platform/middleware"
	"codeflow/internal/platform/migrator"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"syscall"
	"time"

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
	l.Info("System initalized", "env", cfg.Env, "port", cfg.Port)

	port := cfg.Port
	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	m := &metrics.Metrics{}
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewRefreshTokenRepo(db)
	executionRepo := executionPostgres.NewExecutionRepo(db)
	authService := auth.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)
	executionService := execution.NewExecutionService(executionRepo, redisClient)
	executionHandler := execution.NewExecutionHandler(executionService, redisClient, m)

	mux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	protectedMux.HandleFunc("POST /api/v1/executions", executionHandler.Submit)
	protectedMux.HandleFunc("GET /api/v1/executions/{id}", executionHandler.GetByID)
	protectedMux.HandleFunc("GET /api/v1/executions", executionHandler.GetByUser)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /ready", readinessHandler(db, redisClient))
	mux.HandleFunc("GET /metrics", metricsHandler(m, redisClient))

	//User Routes
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

	//Execution Routes
	mux.Handle("/api/v1/executions", middleware.TokenChecker(cfg.JWTSecret)(protectedMux))
	mux.Handle("/api/v1/executions/", middleware.TokenChecker(cfg.JWTSecret)(protectedMux))

	http.Handle("/", middleware.Recovery(middleware.Tracer(middleware.RequestLogger(mux))))

	l.Info("Server starting ", "port", port)

	if err := migrator.Run(cfg.DatabaseURL); err != nil {
		l.Error("Got error in migration", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		l.Info("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			l.Error("shutdown error", "error", err)
		}
	}()

	l.Info("api server starting", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Error("Server failed to start", "error", err)
	}
	l.Info("api server stopped")
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

func metricsHandler(m *metrics.Metrics, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		started, completed, err := m.GetMetrics()

		depth, error := redisClient.LLen(ctx, "executions:queue").Result()
		if error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("couldn't get redis queue length"))
			return
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		output := fmt.Sprintf(
			`# HELP codeflow_executions_started Total executions started
# TYPE codeflow_executions_started counter
codeflow_executions_started %d
# HELP codeflow_executions_completed Total executions completed
# TYPE codeflow_executions_completed counter
codeflow_executions_completed %d
# HELP codeflow_errors Total errors
# TYPE codeflow_errors counter
codeflow_errors %d
# HELP codeflow_queue_depth Current queue depth
# TYPE codeflow_queue_depth gauge
codeflow_queue_depth %d
`, started, completed, err, depth)
		w.Write([]byte(output))
	}
}
