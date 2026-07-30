package service

import (
	"context"
	"fmt"
	"book-management/models"
	"book-management/repository"
)

type ReservationService interface {
	CreateReservation(ctx context.Context, bookID, userID int64) (*models.Reservation, error)
	GetUserReservations(ctx context.Context, userID int64) ([]models.Reservation, error)
	GetReservation(ctx context.Context, id int64) (*models.Reservation, error)
	CancelReservation(ctx context.Context, reservationID, userID int64) error
	FulfillReservation(ctx context.Context, reservationID int64) error
}

type reservationService struct {
	reservationRepo repository.ReservationRepository
	bookRepo        repository.BookRepository
}

func NewReservationService(
	reservationRepo repository.ReservationRepository,
	bookRepo repository.BookRepository,
) ReservationService {
	return &reservationService{
		reservationRepo: reservationRepo,
		bookRepo:        bookRepo,
	}
}

func (s *reservationService) CreateReservation(ctx context.Context, bookID, userID int64) (*models.Reservation, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("book not found")
	}

	if book.AvailableCopies <= 0 {
		return nil, fmt.Errorf("book is not available for reservation")
	}

	hasActive, err := s.reservationRepo.HasActiveReservation(ctx, bookID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing reservation")
	}
	if hasActive {
		return nil, fmt.Errorf("you already have an active reservation for this book")
	}

	reservation := &models.Reservation{
		BookID: bookID,
		UserID: userID,
		Status: "pending",
	}

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	return reservation, nil
}

func (s *reservationService) GetUserReservations(ctx context.Context, userID int64) ([]models.Reservation, error) {
	return s.reservationRepo.GetUserReservations(ctx, userID)
}

func (s *reservationService) GetReservation(ctx context.Context, id int64) (*models.Reservation, error) {
	return s.reservationRepo.GetByID(ctx, id)
}

func (s *reservationService) CancelReservation(ctx context.Context, reservationID, userID int64) error {
	return s.reservationRepo.Cancel(ctx, reservationID, userID)
}

func (s *reservationService) FulfillReservation(ctx context.Context, reservationID int64) error {
	reservation, err := s.reservationRepo.GetByID(ctx, reservationID)
	if err != nil {
		return err
	}

	if reservation.Status != "pending" {
		return fmt.Errorf("only pending reservations can be fulfilled")
	}

	return s.reservationRepo.Fulfill(ctx, reservationID, reservation.BookID)

}