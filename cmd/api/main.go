package main

import (
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"codeflow/internal/platform/middleware"
	"codeflow/internal/platform/migrator"
	"log"
	"net/http"
	"os"
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

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

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
