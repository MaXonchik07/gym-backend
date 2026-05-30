package booking

import (
	"encoding/json"
	"net/http"

	"github.com/MaXonchik07/gym-backend/pkg/middleware"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type Handler struct {
	service Service
	logger  zerolog.Logger
	hub     *Hub
}

func NewHandler(service Service, logger zerolog.Logger, hub *Hub) *Handler {
	return &Handler{service: service, logger: logger, hub: hub}
}

// @Security BearerAuth
// @Summary      Запись на занятие
// @Description  Создаёт бронирование для текущего пользователя
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        request body BookRequest true "Данные записи"
// @Success      201  {object}  models.Booking
// @Failure      400  {string}  string  "неверный запрос"
// @Failure      409  {string}  string  "конфликт"
// @Router       /api/bookings/create [post]
func (h *Handler) BookClass(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	booking, err := h.service.BookClass(r.Context(), claims.UserID, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("booking failed")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "new_booking",
		"booking": booking,
	})

	h.hub.Broadcast(msg)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

// @Security BearerAuth
// @Summary      Мои записи
// @Description  Получает все бронирования для текущего пользователя
// @Tags         bookings
// @Produce      json
// @Success      200  {array}   models.Booking
// @Failure      401  {string}  string  "не авторизован"
// @Router       /api/bookings [get]
func (h *Handler) GetBookings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	bookings, err := h.service.GetUserBookings(r.Context(), claims.UserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get bookings")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bookings)
}

// @Security BearerAuth
// @Summary      Отмена записи
// @Description  Удаляет бронирование по ID
// @Tags         bookings
// @Param        id   query     string  true  "ID бронирования"
// @Success      204  "Запись отменена"
// @Failure      400  {string}  string  "неверный ID"
// @Failure      401  {string}  string  "не авторизован"
// @Router       /api/bookings/cancel [delete]
func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	bookingID := r.URL.Query().Get("id")
	if bookingID == "" {
		http.Error(w, "booking id is required", http.StatusBadRequest)
		return
	}
	if err := h.service.CancelBooking(r.Context(), bookingID, claims.UserID); err != nil {
		h.logger.Error().Err(err).Msg("cancel booking failed")
		http.Error(w, "cancel failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Hub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

func (h *Handler) GetChatUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok || claims.Role != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	users, err := h.service.GetChatUsers(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("get chat users failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) GetChatHistory(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserClaims(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        http.Error(w, "user_id required", http.StatusBadRequest)
        return
    }
    if claims.Role != "admin" && claims.UserID != userID {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }
    msgs, err := h.service.GetRecentMessagesForUser(r.Context(), userID)
    if err != nil {
        h.logger.Error().Err(err).Msg("get chat history failed")
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(msgs)
}