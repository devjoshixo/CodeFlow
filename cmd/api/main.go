package main

import (
	"codeflow/internal/auth"
	"codeflow/internal/auth/postgres"
	"codeflow/internal/execution"
	executionPostgres "codeflow/internal/execution/postgres"
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"codeflow/internal/platform/middleware"
	"codeflow/internal/platform/migrator"
	"context"
	"log"
	"net/http"
	"os"

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

	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewRefreshTokenRepo(db)
	executionRepo := executionPostgres.NewExecutionRepo(db)
	authService := auth.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)
	executionService := execution.NewExecutionService(executionRepo, redisClient)
	executionHandler := execution.NewExecutionHandler(executionService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	//User Routes
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

	//Execution Routes
	mux.HandleFunc("POST /api/v1/executions", executionHandler.Submit)
	mux.HandleFunc("GET /api/v1/executions/{id}", executionHandler.GetByID)
	mux.HandleFunc("GET /api/v1/executions", executionHandler.GetByUser)

	http.Handle("/", middleware.Recovery(middleware.Tracer(middleware.RequestLogger(mux))))

	l.Info("Server starting ", "port", port)

	if err := migrator.Run(cfg.DatabaseURL); err != nil {
		l.Error("Got error in migration", "error", err)
		os.Exit(1)
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		l.Error("Server failed to start", "error", err)
		os.Exit(1)
	}

}
