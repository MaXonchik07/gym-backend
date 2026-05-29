package main

import (
	"context"
	"net/http"

	_ "github.com/MaXonchik07/gym-backend/docs/booking"
	"github.com/MaXonchik07/gym-backend/internal/booking"
	"github.com/MaXonchik07/gym-backend/internal/common"
	"github.com/MaXonchik07/gym-backend/pkg/db"
	"github.com/MaXonchik07/gym-backend/pkg/logger"
	"github.com/MaXonchik07/gym-backend/pkg/middleware"
	"github.com/swaggo/http-swagger"
)

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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {токен}
func main() {
	cfg := common.LoadConfig()
	log := logger.NewLogger(cfg.LogLevel)
	log.Info().Msg("Starting booking service...")
	port := common.GetEnv("BOOKING_SERVER_PORT", "8081")

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer pool.Close()

	msgRepo := booking.NewMessageRepository(pool)
	hub := booking.NewHub(log, msgRepo)
	repo := booking.NewRepository(pool)
	svc := booking.NewService(repo, msgRepo)
	handler := booking.NewHandler(svc, log, hub)

	mux := http.NewServeMux()
	authMux := http.NewServeMux()
	authMux.HandleFunc("/api/bookings", handler.GetBookings)
	authMux.HandleFunc("/api/bookings/create", handler.BookClass)
	authMux.HandleFunc("/api/bookings/cancel", handler.CancelBooking)
	mux.Handle("/api/bookings", middleware.AuthMiddleware(cfg.JWTSecret)(authMux))
	mux.Handle("/api/bookings/create", middleware.AuthMiddleware(cfg.JWTSecret)(authMux))
	mux.Handle("/api/bookings/cancel", middleware.AuthMiddleware(cfg.JWTSecret)(authMux))
	mux.Handle("/api/admin/chat-users", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(handler.GetChatUsers)))
	mux.Handle("/api/admin/chat-history", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(handler.GetChatHistory)))
	mux.HandleFunc("/ws", hub.HandleWebSocket)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	handlerWithCORS := corsMiddleware(mux)

	log.Info().Str("port", port).Msg("Listening")
	if err := http.ListenAndServe(":"+port, handlerWithCORS); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
