package service

import (
	"context"
	"fmt"
	"time"
	"book-management/models"
	"book-management/repository"
)

const (
	maxActiveBorrows = 5  
	loanDays         = 14 
)

type BorrowingService interface {
	IssueBook(ctx context.Context, bookID, userID int64) error
	ReturnBook(ctx context.Context, bookID, userID int64) error
	GetMyBorrowings(ctx context.Context, userID int64) ([]models.Borrowing, error)
	GetOverdueBorrowings(ctx context.Context) ([]models.Borrowing, error)
}

type borrowingService struct {
	borrowingRepo repository.BorrowingRepository
	bookRepo      repository.BookRepository
	fineStrategy  FineStrategy
}

func NewBorrowingService(borrowingRepo repository.BorrowingRepository, bookRepo repository.BookRepository) BorrowingService {
	return &borrowingService{
		borrowingRepo: borrowingRepo,
		bookRepo:      bookRepo,
		fineStrategy:  NewDailyFlatFine(10), 
	}
}

func (s *borrowingService) IssueBook(ctx context.Context, bookID, userID int64) error {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return fmt.Errorf("book not found")
	}
	if book.AvailableCopies <= 0 {
		return fmt.Errorf("book is currently not available")
	}

	hasOverdue, err := s.borrowingRepo.HasOverdueBorrowing(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check overdue status")
	}
	if hasOverdue {
		return fmt.Errorf("you have overdue books; return them before borrowing more")
	}

	activeCount, err := s.borrowingRepo.CountActiveBorrowings(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to count active borrowings")
	}
	if activeCount >= maxActiveBorrows {
		return fmt.Errorf("borrow limit reached (max %d books)", maxActiveBorrows)
	}

	hasActive, err := s.borrowingRepo.HasActiveBorrowing(ctx, bookID, userID)
	if err != nil {
		return fmt.Errorf("failed to check borrowing status")
	}
	if hasActive {
		return fmt.Errorf("you already have an active borrowing for this book")
	}

	borrowing := &models.Borrowing{
		BookID:  bookID,
		UserID:  userID,
		DueDate: time.Now().AddDate(0, 0, loanDays),
		Status:  "borrowed",
	}

	return s.borrowingRepo.IssueBook(ctx, borrowing)
}

func (s *borrowingService) ReturnBook(ctx context.Context, bookID, userID int64) error {
	return s.borrowingRepo.ReturnBook(ctx, bookID, userID)
}

func (s *borrowingService) GetMyBorrowings(ctx context.Context, userID int64) ([]models.Borrowing, error) {
	return s.borrowingRepo.GetMyBorrowings(ctx, userID)
}

func (s *borrowingService) GetOverdueBorrowings(ctx context.Context) ([]models.Borrowing, error) {
	return s.borrowingRepo.GetOverdueBorrowings(ctx)
}