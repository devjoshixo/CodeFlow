package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Gateway_Port string
	DatabaseURL  string
	JWTSecret    string
	RedisURL     string
	Env          string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from system environment")
	}
	cfg := &Config{
		Port:         getenv("PORT", "8080"),
		Gateway_Port: getenv("GATEWAY_PORT", "8081"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		RedisURL:     getenv("REDIS_URL", "redis://localhost:6379"),
		Env:          getenv("ENV", "development"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getenv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
