package booking

import (
	"context"
	"errors"

	"github.com/MaXonchik07/gym-backend/internal/models"
)

type Service interface {
	BookClass(ctx context.Context, userID string, req *BookRequest) (*models.Booking, error)
	GetUserBookings(ctx context.Context, userID string) ([]models.Booking, error)
	CancelBooking(ctx context.Context, bookingID, userID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) BookClass(ctx context.Context, userID string, req *BookRequest) (*models.Booking, error) {
	exists, err := s.repo.IsAlreadyBooked(ctx, userID, req.ClassID, req.Date, req.Time)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("вы уже записаны на это занятие")
	}

	booking := &models.Booking{
		UserID:     userID,
		ClassID:    req.ClassID,
		ClassName:  req.ClassName,
		Instructor: req.Instructor,
		Date:       req.Date,
		Time:       req.Time,
	}
	if err := s.repo.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *service) GetUserBookings(ctx context.Context, userID string) ([]models.Booking, error) {
	return s.repo.GetBookingsByUser(ctx, userID)
}

func (s *service) CancelBooking(ctx context.Context, bookingID, userID string) error {
	return s.repo.CancelBooking(ctx, bookingID, userID)
}

type BookRequest struct {
	ClassID    string `json:"class_id"`
	ClassName  string `json:"class_name"`
	Instructor string `json:"instructor"`
	Date       string `json:"date"`
	Time       string `json:"time"`
}
