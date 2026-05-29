package main

import (
	"context"
	"net/http"

	_ "github.com/MaXonchik07/gym-backend/docs/auth"
	"github.com/MaXonchik07/gym-backend/internal/auth"
	"github.com/MaXonchik07/gym-backend/internal/common"
	"github.com/MaXonchik07/gym-backend/pkg/db"
	"github.com/MaXonchik07/gym-backend/pkg/logger"
	"github.com/MaXonchik07/gym-backend/pkg/middleware"
	"github.com/swaggo/http-swagger"
)

// @title           Gym Auth Service API
// @version         1.0
// @description     API для регистрации, входа и управления пользователями.
// @host            localhost:8080
// @BasePath        /api/auth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {токен}

func main() {
	cfg := common.LoadConfig()
	log := logger.NewLogger(cfg.LogLevel)
	log.Info().Msg("Starting auth service...")
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
	mux.Handle("/api/auth/me", middleware.AuthMiddleware(cfg.JWTSecret)(
		http.HandlerFunc(handler.Me),
	))
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/api/auth/profile", middleware.AuthMiddleware(cfg.JWTSecret)(
		http.HandlerFunc(handler.UpdateProfile),
	))
	mux.Handle("/api/auth/membership", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(handler.UpdateMembership)))

	handlerWithCORS := corsMiddleware(mux)
	log.Info().Str("port", port).Msg("Listening")
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
