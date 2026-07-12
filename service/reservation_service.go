package service

import (
	"context"
	"database/sql"
	"errors"

	"book-management/models"
	"book-management/repository"
)

type ReservationService struct {
	bookRepo        *repository.BookRepository
	reservationRepo *repository.ReservationRepository
}

func NewReservationService(db *sql.DB) *ReservationService {
	return &ReservationService{
		bookRepo:        repository.NewBookRepository(db),
		reservationRepo: repository.NewReservationRepository(db),
	}
}

func (s *ReservationService) Create(ctx context.Context, bookID, userID int64) (models.Reservation, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return models.Reservation{}, errors.New("book not found")
	}

	if book.AvailableCopies > 0 {
		return models.Reservation{}, errors.New("book is available, no need to reserve")
	}
	
	return s.reservationRepo.Create(ctx, models.Reservation{
		BookID: bookID,
		UserID: userID,
	})
}