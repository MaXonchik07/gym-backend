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
	GetChatUsers(ctx context.Context) ([]string, error)
	GetMessagesForUser(ctx context.Context, userID string) ([]models.Message, error)
}

type service struct {
	repo    Repository
	msgRepo MessageRepository
}

func NewService(repo Repository, msgRepo MessageRepository) Service {
	return &service{repo: repo, msgRepo: msgRepo}
}

func (s *service) BookClass(ctx context.Context, userID string, req *BookRequest) (*models.Booking, error) {
	exists, err := s.repo.IsAlreadyBooked(ctx, userID, req.ClassID, req.Date, req.Time)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("Вы уже записаны на это занятие")
	}

	current, err := s.repo.CountBookingsForSlot(ctx, req.ClassID, req.Date, req.Time)
	if err != nil {
		return nil, err
	}
	if req.Capacity > 0 && current >= req.Capacity {
		return nil, errors.New("Нет доступных мест")
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

func (s *service) GetChatUsers(ctx context.Context) ([]string, error) {
	return s.msgRepo.GetChatUsers(ctx)
}

func (s *service) GetMessagesForUser(ctx context.Context, userID string) ([]models.Message, error) {
	return s.msgRepo.GetRecentMessagesForUser(ctx, userID, 50)
}

type BookRequest struct {
	ClassID    string `json:"class_id"`
	ClassName  string `json:"class_name"`
	Instructor string `json:"instructor"`
	Date       string `json:"date"`
	Time       string `json:"time"`
	Capacity   int    `json:"capacity"`
}
