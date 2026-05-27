package booking

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/MaXonchik07/gym-backend/internal/models"
)

func TestRepoCreateBooking(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewRepository(mock)

	mock.ExpectQuery(`INSERT INTO bookings`).
		WithArgs("user-1", "yoga", "Yoga", "Anna", "2026-06-01", "07:00").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).
			AddRow("booking-1", time.Now()))

	b := &models.Booking{
		UserID:     "user-1",
		ClassID:    "yoga",
		ClassName:  "Yoga",
		Instructor: "Anna",
		Date:       "2026-06-01",
		Time:       "07:00",
	}
	err = repo.CreateBooking(context.Background(), b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ID != "booking-1" {
		t.Errorf("expected booking-1, got %s", b.ID)
	}
}

func TestRepoGetBookingsByUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewRepository(mock)

	mock.ExpectQuery(`SELECT (.+) FROM bookings`).
		WithArgs("user-1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "class_id", "class_name", "instructor", "date", "time", "created_at"}).
			AddRow("b1", "user-1", "yoga", "Yoga", "Anna", "2026-06-01", "07:00", time.Now()).
			AddRow("b2", "user-1", "boxing", "Boxing", "Mike", "2026-06-02", "08:00", time.Now()))

	bookings, err := repo.GetBookingsByUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bookings) != 2 {
		t.Errorf("expected 2 bookings, got %d", len(bookings))
	}
}

func TestRepoCancelBooking(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewRepository(mock)

	mock.ExpectExec(`DELETE FROM bookings`).
		WithArgs("b1", "user-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.CancelBooking(context.Background(), "b1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRepoIsAlreadyBooked(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewRepository(mock)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("user-1", "yoga", "2026-06-01", "07:00").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.IsAlreadyBooked(context.Background(), "user-1", "yoga", "2026-06-01", "07:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true, got false")
	}
}