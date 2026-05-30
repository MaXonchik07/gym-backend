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
	b.ID = "mock-id-123"
	m.bookings = append(m.bookings, *b)
	return nil
}
func (m *mockRepo) GetBookingsByUser(ctx context.Context, userID string) ([]models.Booking, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []models.Booking
	for _, b := range m.bookings {
		if b.UserID == userID {
			res = append(res, b)
		}
	}
	return res, nil
}
func (m *mockRepo) CancelBooking(ctx context.Context, bookingID, userID string) error {
	if m.cancelErr != nil {
		return m.cancelErr
	}
	for i, b := range m.bookings {
		if b.ID == bookingID && b.UserID == userID {
			m.bookings = append(m.bookings[:i], m.bookings[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}
func (m *mockRepo) IsAlreadyBooked(ctx context.Context, userID, classID, date, time string) (bool, error) {
	if m.isBookedErr != nil {
		return false, m.isBookedErr
	}
	return m.isBookedResult, nil
}
func (m *mockRepo) CountBookingsForSlot(ctx context.Context, classID, date, time string) (int, error) {
	if m.countBookingsFn != nil {
		return m.countBookingsFn(ctx, classID, date, time)
	}
	return 0, nil
}

type mockMessageRepo struct {
	getConversationFn          func(ctx context.Context, userID string, limit int) ([]models.Message, error)
	getChatUsersFn             func(ctx context.Context) ([]string, error)
	getRecentMessagesForUserFn func(ctx context.Context, userID string, limit int) ([]models.Message, error)
	getUserNameByIDFn          func(ctx context.Context, userID string) (string, string, error)
}

func (m *mockMessageRepo) GetUserNameByID(ctx context.Context, userID string) (string, string, error) {
    if m.getUserNameByIDFn != nil {
        return m.getUserNameByIDFn(ctx, userID)
    }
    return "Имя", "Фамилия", nil
}

func (m *mockMessageRepo) SaveMessage(ctx context.Context, msg *models.Message) error { return nil }

func (m *mockMessageRepo) GetRecentMessagesForUser(ctx context.Context, userID string, limit int) ([]models.Message, error) {
	if m.getRecentMessagesForUserFn != nil {
		return m.getRecentMessagesForUserFn(ctx, userID, limit)
	}
	return nil, nil
}
func (m *mockMessageRepo) GetChatUsers(ctx context.Context) ([]string, error) {
	if m.getChatUsersFn != nil {
		return m.getChatUsersFn(ctx)
	}
	return nil, nil
}
func (m *mockMessageRepo) GetConversation(ctx context.Context, userID string, limit int) ([]models.Message, error) {
	if m.getConversationFn != nil {
		return m.getConversationFn(ctx, userID, limit)
	}
	return nil, nil
}

func TestBookClass_Success(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockMessageRepo{})
	req := &BookRequest{ClassID: "yoga", ClassName: "Yoga Flow", Instructor: "Anna", Date: "2026-06-01", Time: "07:00"}
	b, err := svc.BookClass(context.Background(), "u1", req)
	if err != nil || b.ID != "mock-id-123" {
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

func TestGetUserBookings(t *testing.T) {
	repo := &mockRepo{bookings: []models.Booking{{ID: "1", UserID: "u1"}, {ID: "2", UserID: "u2"}}}
	svc := NewService(repo, &mockMessageRepo{})
	b, _ := svc.GetUserBookings(context.Background(), "u1")
	if len(b) != 1 {
		t.Errorf("expected 1, got %d", len(b))
	}
}

func TestCancelBooking(t *testing.T) {
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

func TestBookClass_AfterCancel(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockMessageRepo{})
	req := &BookRequest{ClassID: "yoga", ClassName: "Yoga", Instructor: "Anna", Date: "2026-06-01", Time: "07:00"}
	b1, _ := svc.BookClass(context.Background(), "u1", req)
	svc.CancelBooking(context.Background(), b1.ID, "u1")
	repo.isBookedResult = false
	b2, err := svc.BookClass(context.Background(), "u1", req)
	if err != nil || b2.ID == "" {
		t.Fatalf("second booking failed: %v", err)
	}
}

func TestGetChatUsers(t *testing.T) {
	msgRepo := &mockMessageRepo{getChatUsersFn: func(ctx context.Context) ([]string, error) { return []string{"u1", "u2"}, nil }}
	svc := NewService(&mockRepo{}, msgRepo)
	u, _ := svc.GetChatUsers(context.Background())
	if len(u) != 2 {
		t.Errorf("expected 2, got %d", len(u))
	}
}

func TestGetConversation(t *testing.T) {
	msgRepo := &mockMessageRepo{getConversationFn: func(ctx context.Context, userID string, limit int) ([]models.Message, error) {
		return []models.Message{{ID: "m1", Content: "Hello"}}, nil
	}}
	svc := NewService(&mockRepo{}, msgRepo)
	msgs, _ := svc.GetConversation(context.Background(), "u1")
	if len(msgs) != 1 {
		t.Errorf("expected 1, got %d", len(msgs))
	}
}
