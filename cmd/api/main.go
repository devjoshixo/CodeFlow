package main

import (
	"codeflow/internal/platform/config"
	"codeflow/internal/platform/logger"
	"codeflow/internal/platform/migrator"
	"fmt"
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

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status": "ok"}`)
	})

	l.Info("Server started on the ", "port", port)

	if err := migrator.Run(cfg.DatabaseURL); err != nil {
		l.Error("Got error in migration", "error", err)
		os.Exit(1)
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		l.Error("Server failed to start", "error", err)
		os.Exit(1)
	}

}
