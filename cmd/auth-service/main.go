package main

import (
	"context"
	"net/http"

	"github.com/MaXonchik07/gym-backend/internal/auth"
	"github.com/MaXonchik07/gym-backend/internal/common"
	"github.com/MaXonchik07/gym-backend/pkg/db"
	"github.com/MaXonchik07/gym-backend/pkg/logger"
	"github.com/swaggo/http-swagger"
	_ "github.com/MaXonchik07/gym-backend/docs/auth"
)

// @title           Gym Auth Service API
// @version         1.0
// @description     API для регистрации, входа и управления пользователями.
// @host            localhost:8080
// @BasePath        /api/auth

func main() {
	cfg := common.LoadConfig()
	log := logger.NewLogger(cfg.LogLevel)
	log.Info().Msg("Starting auth service...")

	// Порт для auth-сервиса (по умолчанию 8080)
	port := common.GetEnv("AUTH_SERVER_PORT", "8080")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer pool.Close()

	repo := auth.NewRepository(pool)
	svc := auth.NewService(repo, cfg.JWTSecret)
	handler := auth.NewHandler(svc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/register", handler.Register)
	mux.HandleFunc("/api/auth/login", handler.Login)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Info().Str("port", port).Msg("Listening")
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
