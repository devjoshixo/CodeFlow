package main

import (
	"codeflow/internal/auth"
	"codeflow/internal/auth/postgres"
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"codeflow/internal/platform/middleware"
	"codeflow/internal/platform/migrator"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
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

	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewRefreshTokenRepo(db)
	authService := auth.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

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
