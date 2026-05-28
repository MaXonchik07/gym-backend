package booking

import (
	"context"
	"errors"
	"testing"
	"github.com/MaXonchik07/gym-backend/internal/models"
)

type mockRepo struct {
	bookings        []models.Booking
	createErr       error
	getErr          error
	cancelErr       error
	isBookedErr     error
	isBookedResult  bool
	countBookingsFn func(ctx context.Context, classID, date, time string) (int, error)
}

func (m *mockRepo) CreateBooking(ctx context.Context, b *models.Booking) error {
	if m.createErr != nil {
		return m.createErr
	}
	b.ID = "mock-id"
	m.bookings = append(m.bookings, *b)
	return nil
}
func (m *mockRepo) GetBookingsByUser(ctx context.Context, uid string) ([]models.Booking, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []models.Booking
	for _, b := range m.bookings {
		if b.UserID == uid {
			res = append(res, b)
		}
	}
	return res, nil
}
func (m *mockRepo) CancelBooking(ctx context.Context, bid, uid string) error {
	if m.cancelErr != nil {
		return m.cancelErr
	}
	for i, b := range m.bookings {
		if b.ID == bid && b.UserID == uid {
			m.bookings = append(m.bookings[:i], m.bookings[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}
func (m *mockRepo) IsAlreadyBooked(ctx context.Context, uid, cid, date, time string) (bool, error) {
	if m.isBookedErr != nil {
		return false, m.isBookedErr
	}
	return m.isBookedResult, nil
}
func (m *mockRepo) CountBookingsForSlot(ctx context.Context, cid, date, time string) (int, error) {
	if m.countBookingsFn != nil {
		return m.countBookingsFn(ctx, cid, date, time)
	}
	return 0, nil
}

type mockMessageRepo struct {
	getRecentFn func(ctx context.Context, uid string, limit int) ([]models.Message, error)
	getUsersFn  func(ctx context.Context) ([]string, error)
}

func (m *mockMessageRepo) SaveMessage(ctx context.Context, msg *models.Message) error { return nil }
func (m *mockMessageRepo) GetRecentMessagesForUser(ctx context.Context, uid string, limit int) ([]models.Message, error) {
	if m.getRecentFn != nil {
		return m.getRecentFn(ctx, uid, limit)
	}
	return nil, nil
}
func (m *mockMessageRepo) GetChatUsers(ctx context.Context) ([]string, error) {
	if m.getUsersFn != nil {
		return m.getUsersFn(ctx)
	}
	return nil, nil
}

func TestBookClass_Success(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockMessageRepo{})
	b, err := svc.BookClass(context.Background(), "u1", &BookRequest{ClassID: "yoga", ClassName: "Y", Instructor: "A", Date: "2026-06-01", Time: "07:00"})
	if err != nil || b.ID != "mock-id" {
		t.Fatalf("unexpected: %v %s", err, b.ID)
	}
}

func TestBookClass_AlreadyBooked(t *testing.T) {
	svc := NewService(&mockRepo{isBookedResult: true}, &mockMessageRepo{})
	_, err := svc.BookClass(context.Background(), "u1", &BookRequest{ClassID: "yoga", Date: "2026-06-01", Time: "07:00"})
	if err == nil || err.Error() != "вы уже записаны на это занятие" {
		t.Errorf("unexpected: %v", err)
	}
}

func TestBookClass_CapacityExceeded(t *testing.T) {
	repo := &mockRepo{countBookingsFn: func(ctx context.Context, cid, d, t string) (int, error) { return 20, nil }}
	svc := NewService(repo, &mockMessageRepo{})
	_, err := svc.BookClass(context.Background(), "u1", &BookRequest{ClassID: "yoga", Date: "2026-06-01", Time: "07:00", Capacity: 20})
	if err == nil || err.Error() != "нет доступных мест" {
		t.Errorf("expected capacity error, got %v", err)
	}
}

func TestBookClass_CreateError(t *testing.T) {
	svc := NewService(&mockRepo{createErr: errors.New("db error")}, &mockMessageRepo{})
	_, err := svc.BookClass(context.Background(), "u1", &BookRequest{ClassID: "yoga", Date: "2026-06-01", Time: "07:00"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetUserBookings_Success(t *testing.T) {
	repo := &mockRepo{bookings: []models.Booking{{ID: "1", UserID: "u1"}, {ID: "2", UserID: "u2"}}}
	svc := NewService(repo, &mockMessageRepo{})
	b, _ := svc.GetUserBookings(context.Background(), "u1")
	if len(b) != 1 {
		t.Errorf("expected 1, got %d", len(b))
	}
}

func TestGetUserBookings_Error(t *testing.T) {
	svc := NewService(&mockRepo{getErr: errors.New("db error")}, &mockMessageRepo{})
	_, err := svc.GetUserBookings(context.Background(), "u1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCancelBooking_Success(t *testing.T) {
	repo := &mockRepo{bookings: []models.Booking{{ID: "1", UserID: "u1"}}}
	svc := NewService(repo, &mockMessageRepo{})
	err := svc.CancelBooking(context.Background(), "1", "u1")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(repo.bookings) != 0 {
		t.Error("not deleted")
	}
}

func TestCancelBooking_NotFound(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockMessageRepo{})
	err := svc.CancelBooking(context.Background(), "1", "u1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCancelBooking_Error(t *testing.T) {
	svc := NewService(&mockRepo{cancelErr: errors.New("db error")}, &mockMessageRepo{})
	err := svc.CancelBooking(context.Background(), "1", "u1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetChatUsers(t *testing.T) {
	msgRepo := &mockMessageRepo{getUsersFn: func(ctx context.Context) ([]string, error) { return []string{"u1", "u2"}, nil }}
	svc := NewService(&mockRepo{}, msgRepo)
	u, _ := svc.GetChatUsers(context.Background())
	if len(u) != 2 {
		t.Errorf("expected 2, got %d", len(u))
	}
}

func TestGetMessagesForUser(t *testing.T) {
	msgRepo := &mockMessageRepo{getRecentFn: func(ctx context.Context, uid string, limit int) ([]models.Message, error) {
		return []models.Message{{ID: "m1", Content: "Hi"}}, nil
	}}
	svc := NewService(&mockRepo{}, msgRepo)
	msgs, _ := svc.GetMessagesForUser(context.Background(), "u1")
	if len(msgs) != 1 {
		t.Errorf("expected 1, got %d", len(msgs))
	}
}

func TestBookClass_CountBookingsError(t *testing.T) {
    repo := &mockRepo{
        countBookingsFn: func(ctx context.Context, cid, d, t string) (int, error) {
            return 0, errors.New("db error")
        },
    }
    svc := NewService(repo, &mockMessageRepo{})
    _, err := svc.BookClass(context.Background(), "u1", &BookRequest{ClassID: "yoga", Date: "2026-06-01", Time: "07:00"})
    if err == nil {
        t.Error("expected error from CountBookingsForSlot")
    }
}

