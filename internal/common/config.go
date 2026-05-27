package common

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	LogLevel    string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL: GetEnv("DATABASE_URL", "postgres://gym:gym@localhost:5432/gymdb?sslmode=disable"),
		JWTSecret:   GetEnv("JWT_SECRET", "super-secret-key"),
		LogLevel:    GetEnv("LOG_LEVEL", "debug"),
	}
}

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}