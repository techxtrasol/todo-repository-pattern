package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppPort     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://todo:todo123@localhost:5432/todo_db?sslmode=disable"),
		AppPort:     getEnv("APP_PORT", "8080"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
