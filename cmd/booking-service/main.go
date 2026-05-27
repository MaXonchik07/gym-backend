package main

import (
	"context"
	"net/http"

	"github.com/MaXonchik07/gym-backend/internal/booking"
	"github.com/MaXonchik07/gym-backend/internal/common"
	"github.com/MaXonchik07/gym-backend/pkg/db"
	"github.com/MaXonchik07/gym-backend/pkg/logger"
	"github.com/MaXonchik07/gym-backend/pkg/middleware"
	"github.com/swaggo/http-swagger"
	_ "github.com/MaXonchik07/gym-backend/docs/booking"
)

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
	svc := booking.NewService(repo)
	handler := booking.NewHandler(svc, log, hub)

	mux := http.NewServeMux()
	authMux := http.NewServeMux()
	authMux.HandleFunc("/api/bookings", handler.GetBookings)
	authMux.HandleFunc("/api/bookings/create", handler.BookClass)
	authMux.HandleFunc("/api/bookings/cancel", handler.CancelBooking)
	mux.Handle("/api/bookings", middleware.AuthMiddleware(cfg.JWTSecret)(authMux))
	mux.Handle("/api/bookings/create", middleware.AuthMiddleware(cfg.JWTSecret)(authMux))
	mux.Handle("/api/bookings/cancel", middleware.AuthMiddleware(cfg.JWTSecret)(authMux))

	mux.HandleFunc("/ws", hub.HandleWebSocket)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Info().Str("port", port).Msg("Listening")
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}