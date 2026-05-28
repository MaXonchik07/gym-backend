package booking

import (
	"context"
	"errors"
	"github.com/MaXonchik07/gym-backend/internal/models"
	"testing"
	"time"
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

func (m *mockRepo) CountBookingsForSlot(ctx context.Context, classID, date, time string) (int, error) {
	if m.countBookingsFn != nil {
		return m.countBookingsFn(ctx, classID, date, time)
	}
	return 0, nil
}

func (m *mockRepo) CreateBooking(ctx context.Context, b *models.Booking) error {
	if m.createErr != nil {
		return m.createErr
	}
	b.ID = "mock-id-123"
	b.CreatedAt = time.Now()
	m.bookings = append(m.bookings, *b)
	return nil
}

func (m *mockRepo) GetBookingsByUser(ctx context.Context, userID string) ([]models.Booking, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var result []models.Booking
	for _, b := range m.bookings {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
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

func TestBookClass_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)
	req := &BookRequest{
		ClassID:    "yoga",
		ClassName:  "Yoga Flow",
		Instructor: "Anna",
		Date:       "2026-06-01",
		Time:       "07:00",
	}
	booking, err := svc.BookClass(context.Background(), "user1", req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if booking.ID != "mock-id-123" {
		t.Errorf("expected mock-id-123, got %s", booking.ID)
	}
}

func TestBookClass_AlreadyBooked(t *testing.T) {
	repo := &mockRepo{
		isBookedResult: true,
	}
	svc := NewService(repo)
	req := &BookRequest{
		ClassID: "yoga",
		Date:    "2026-06-01",
		Time:    "07:00",
	}
	_, err := svc.BookClass(context.Background(), "user1", req)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != "вы уже записаны на это занятие" {
		t.Errorf("expected duplicate error, got %v", err)
	}
}

func TestGetUserBookings(t *testing.T) {
	repo := &mockRepo{
		bookings: []models.Booking{
			{ID: "1", UserID: "user1", ClassName: "Yoga"},
			{ID: "2", UserID: "user2", ClassName: "Boxing"},
		},
	}
	svc := NewService(repo)
	bookings, err := svc.GetUserBookings(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("expected 1 booking, got %d", len(bookings))
	}
	if bookings[0].ClassName != "Yoga" {
		t.Errorf("expected Yoga, got %s", bookings[0].ClassName)
	}
}

func TestCancelBooking(t *testing.T) {
	repo := &mockRepo{
		bookings: []models.Booking{
			{ID: "1", UserID: "user1"},
		},
	}
	svc := NewService(repo)
	err := svc.CancelBooking(context.Background(), "1", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.bookings) != 0 {
		t.Error("expected booking to be deleted")
	}
}
func TestCancelBooking_NotFound(t *testing.T) {
	repo := &mockRepo{
		bookings: []models.Booking{
			{ID: "1", UserID: "user1"},
		},
	}
	svc := NewService(repo)
	err := svc.CancelBooking(context.Background(), "non-existent", "user1")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != "not found" {
		t.Errorf("expected 'not found', got %v", err)
	}
}

func TestBookClass_AfterCancel(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)
	req := &BookRequest{
		ClassID:    "yoga",
		ClassName:  "Yoga",
		Instructor: "Anna",
		Date:       "2026-06-01",
		Time:       "07:00",
	}
	booking, err := svc.BookClass(context.Background(), "user1", req)
	if err != nil {
		t.Fatalf("first booking failed: %v", err)
	}
	err = svc.CancelBooking(context.Background(), booking.ID, "user1")
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	repo.isBookedResult = false
	booking2, err := svc.BookClass(context.Background(), "user1", req)
	if err != nil {
		t.Fatalf("second booking after cancel failed: %v", err)
	}
	if booking2.ID == "" {
		t.Error("expected new booking ID")
	}
}
