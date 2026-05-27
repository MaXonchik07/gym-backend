package booking

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/MaXonchik07/gym-backend/internal/models"
)

type Repository interface {
	CreateBooking(ctx context.Context, booking *models.Booking) error
	GetBookingsByUser(ctx context.Context, userID string) ([]models.Booking, error)
	CancelBooking(ctx context.Context, bookingID, userID string) error
	IsAlreadyBooked(ctx context.Context, userID, classID, date, time string) (bool, error)
}

type DBPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type repository struct {
	pool DBPool
}

func NewRepository(pool DBPool) Repository {
	return &repository{pool: pool}
}

func (r *repository) CreateBooking(ctx context.Context, b *models.Booking) error {
	query := `
		INSERT INTO bookings (user_id, class_id, class_name, instructor, date, time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		b.UserID, b.ClassID, b.ClassName, b.Instructor, b.Date, b.Time,
	).Scan(&b.ID, &b.CreatedAt)
}

func (r *repository) GetBookingsByUser(ctx context.Context, userID string) ([]models.Booking, error) {
	query := `
		SELECT id, user_id, class_id, class_name, instructor, date::text, time, created_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY date, time
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.ClassID, &b.ClassName, &b.Instructor, &b.Date, &b.Time, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

func (r *repository) CancelBooking(ctx context.Context, bookingID, userID string) error {
	query := `DELETE FROM bookings WHERE id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, bookingID, userID)
	return err
}

func (r *repository) IsAlreadyBooked(ctx context.Context, userID, classID, date, time string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM bookings WHERE user_id = $1 AND class_id = $2 AND date = $3 AND time = $4)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, classID, date, time).Scan(&exists)
	return exists, err
}
