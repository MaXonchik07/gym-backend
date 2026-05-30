package booking

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaXonchik07/gym-backend/internal/models"
	"github.com/MaXonchik07/gym-backend/pkg/jwt"
	"github.com/MaXonchik07/gym-backend/pkg/middleware"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type mockBookingService struct {
	bookFn              func(ctx context.Context, userID string, req *BookRequest) (*models.Booking, error)
	getFn               func(ctx context.Context, userID string) ([]models.Booking, error)
	cancelFn            func(ctx context.Context, bookingID, userID string) error
	getChatFn           func(ctx context.Context) ([]string, error)
	getConvFn           func(ctx context.Context, userID string) ([]models.Message, error)
	getMessagesFn       func(ctx context.Context, userID string) ([]models.Message, error)
	getRecentMessagesFn func(ctx context.Context, userID string) ([]models.Message, error)
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
func (m *mockBookingService) GetChatUsers(ctx context.Context) ([]string, error) {
	if m.getChatFn != nil {
		return m.getChatFn(ctx)
	}
	return nil, nil
}
func (m *mockBookingService) GetConversation(ctx context.Context, userID string) ([]models.Message, error) {
	if m.getConvFn != nil {
		return m.getConvFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockBookingService) GetMessagesForUser(ctx context.Context, userID string) ([]models.Message, error) {
	if m.getMessagesFn != nil {
		return m.getMessagesFn(ctx, userID)
	}
	return nil, nil
}

func newTestHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]string), logger: zerolog.Nop()}
}

func TestHandler_BookClass(t *testing.T) {
	svc := &mockBookingService{bookFn: func(ctx context.Context, uid string, req *BookRequest) (*models.Booking, error) {
		return &models.Booking{ID: "b1"}, nil
	}}
	handler := NewHandler(svc, zerolog.Nop(), newTestHub())
	body := `{"class_id":"yoga","class_name":"Y","instructor":"A","date":"2026-06-01","time":"07:00"}`
	req := httptest.NewRequest(http.MethodPost, "/bookings/create", bytes.NewBufferString(body))
	claims := &jwt.Claims{UserID: "u1"}
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
	svc := &mockBookingService{getFn: func(ctx context.Context, uid string) ([]models.Booking, error) {
		return []models.Booking{{ID: "b1"}}, nil
	}}
	handler := NewHandler(svc, zerolog.Nop(), newTestHub())
	req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
	claims := &jwt.Claims{UserID: "u1"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.GetBookings(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandler_CancelBooking(t *testing.T) {
	svc := &mockBookingService{cancelFn: func(ctx context.Context, bid, uid string) error { return nil }}
	handler := NewHandler(svc, zerolog.Nop(), newTestHub())
	req := httptest.NewRequest(http.MethodDelete, "/bookings/cancel?id=b1", nil)
	claims := &jwt.Claims{UserID: "u1"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.CancelBooking(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func (m *mockBookingService) GetRecentMessagesForUser(ctx context.Context, userID string) ([]models.Message, error) {
	if m.getRecentMessagesFn != nil {
		return m.getRecentMessagesFn(ctx, userID)
	}
	return nil, nil
}
