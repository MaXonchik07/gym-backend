package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/MaXonchik07/gym-backend/internal/models"
	"github.com/MaXonchik07/gym-backend/pkg/jwt"
	"github.com/MaXonchik07/gym-backend/pkg/middleware"
)

type mockBookingService struct {
	bookFn   func(ctx context.Context, userID string, req *BookRequest) (*models.Booking, error)
	getFn    func(ctx context.Context, userID string) ([]models.Booking, error)
	cancelFn func(ctx context.Context, bookingID, userID string) error
}

func (m *mockBookingService) BookClass(ctx context.Context, userID string, req *BookRequest) (*models.Booking, error) {
	return m.bookFn(ctx, userID, req)
}
func (m *mockBookingService) GetUserBookings(ctx context.Context, userID string) ([]models.Booking, error) {
	return m.getFn(ctx, userID)
}
func (m *mockBookingService) CancelBooking(ctx context.Context, bookingID, userID string) error {
	return m.cancelFn(ctx, bookingID, userID)
}

func newTestHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
		logger:  zerolog.Nop(),
	}
}

// Тесты
func TestHandler_BookClass(t *testing.T) {
	svc := &mockBookingService{
		bookFn: func(ctx context.Context, userID string, req *BookRequest) (*models.Booking, error) {
			return &models.Booking{ID: "b1", ClassName: req.ClassName}, nil
		},
	}
	handler := NewHandler(svc, zerolog.Nop(), newTestHub())

	body := `{"class_id":"yoga","class_name":"Yoga","instructor":"Anna","date":"2026-06-01","time":"07:00"}`
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(body))
	claims := &jwt.Claims{UserID: "user1"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.BookClass(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_GetBookings(t *testing.T) {
	svc := &mockBookingService{
		getFn: func(ctx context.Context, userID string) ([]models.Booking, error) {
			return []models.Booking{{ID: "b1"}}, nil
		},
	}
	handler := NewHandler(svc, zerolog.Nop(), newTestHub())

	req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
	claims := &jwt.Claims{UserID: "user1"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.GetBookings(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CancelBooking(t *testing.T) {
	svc := &mockBookingService{
		cancelFn: func(ctx context.Context, bookingID, userID string) error {
			if bookingID == "b1" && userID == "user1" {
				return nil
			}
			return errors.New("not found")
		},
	}
	handler := NewHandler(svc, zerolog.Nop(), newTestHub())

	req := httptest.NewRequest(http.MethodDelete, "/bookings/cancel?id=b1", nil)
	claims := &jwt.Claims{UserID: "user1"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CancelBooking(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_GetBookings_Empty(t *testing.T) {
	svc := &mockBookingService{
		getFn: func(ctx context.Context, userID string) ([]models.Booking, error) {
			return []models.Booking{}, nil
		},
	}
	handler := NewHandler(svc, zerolog.Nop(), newTestHub())

	req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
	claims := &jwt.Claims{UserID: "user1"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.GetBookings(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var bookings []models.Booking
	json.NewDecoder(rec.Body).Decode(&bookings)
	if len(bookings) != 0 {
		t.Errorf("expected 0 bookings, got %d", len(bookings))
	}
}