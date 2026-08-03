package service

import (
	"book-management/models"
	"book-management/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBorrowingRepository struct {
	mock.Mock
}

func (m *MockBorrowingRepository) IssueBook(ctx context.Context, borrowing *models.Borrowing) error {
	args := m.Called(ctx, borrowing)
	return args.Error(0)
}

func (m *MockBorrowingRepository) ReturnBook(ctx context.Context, bookID, userID int64) error {
	args := m.Called(ctx, bookID, userID)
	return args.Error(0)
}

func (m *MockBorrowingRepository) GetMyBorrowings(ctx context.Context, userID int64) ([]models.Borrowing, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Borrowing), args.Error(1)
}

func (m *MockBorrowingRepository) GetOverdueBorrowings(ctx context.Context) ([]models.Borrowing, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Borrowing), args.Error(1)
}

func (m *MockBorrowingRepository) HasActiveBorrowing(ctx context.Context, bookID, userID int64) (bool, error) {
	args := m.Called(ctx, bookID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBorrowingRepository) CountActiveBorrowings(ctx context.Context, userID int64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockBorrowingRepository) HasOverdueBorrowing(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

var _ repository.BorrowingRepository = (*MockBorrowingRepository)(nil)

type MockBookRepoForBorrowing struct {
	mock.Mock
}

func (m *MockBookRepoForBorrowing) GetByID(ctx context.Context, id int64) (*models.Book, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookRepoForBorrowing) List(ctx context.Context, filter models.BookFilter) ([]models.Book, error) {
	return nil, nil
}
func (m *MockBookRepoForBorrowing) Create(ctx context.Context, book *models.Book) error { return nil }
func (m *MockBookRepoForBorrowing) Update(ctx context.Context, book *models.Book) error { return nil }
func (m *MockBookRepoForBorrowing) Delete(ctx context.Context, id int64) error          { return nil }
func (m *MockBookRepoForBorrowing) UpdateAvailability(ctx context.Context, bookID int64, delta int) error {
	return nil
}
func (m *MockBookRepoForBorrowing) CountByCategory(ctx context.Context, categoryID int64) (int, error) {
	return 0, nil
}
func (m *MockBookRepoForBorrowing) UpdateBookLink(ctx context.Context, bookID int64, url, bookType string) error {
	return nil
}

var _ repository.BookRepository = (*MockBookRepoForBorrowing)(nil)

func availableBook(id int64) *models.Book {
	return &models.Book{
		ID:              id,
		Title:           "Test Book",
		AvailableCopies: 3,
		Quantity:        5,
		Status:          "available",
	}
}

func unavailableBook(id int64) *models.Book {
	return &models.Book{
		ID:              id,
		Title:           "Test Book",
		AvailableCopies: 0,
		Quantity:        5,
		Status:          "borrowed",
	}
}

func TestIssueBook_Success(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	bookID, userID := int64(1), int64(10)

	bookRepo.On("GetByID", ctx, bookID).Return(availableBook(bookID), nil).Once()
	borrowRepo.On("HasOverdueBorrowing", ctx, userID).Return(false, nil).Once()
	borrowRepo.On("CountActiveBorrowings", ctx, userID).Return(2, nil).Once()
	borrowRepo.On("HasActiveBorrowing", ctx, bookID, userID).Return(false, nil).Once()
	borrowRepo.On("IssueBook", ctx, mock.MatchedBy(func(b *models.Borrowing) bool {
		return b.BookID == bookID &&
			b.UserID == userID &&
			b.Status == "borrowed" &&
			b.DueDate.After(time.Now().AddDate(0, 0, loanDays-1))
	})).Return(nil).Once()

	err := svc.IssueBook(ctx, bookID, userID)

	assert.NoError(t, err)
	borrowRepo.AssertExpectations(t)
	bookRepo.AssertExpectations(t)
}

func TestIssueBook_BookNotFound(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	bookRepo.On("GetByID", ctx, int64(99)).Return(nil, errors.New("not found")).Once()

	err := svc.IssueBook(ctx, 99, 10)

	assert.EqualError(t, err, "book not found")
	borrowRepo.AssertNotCalled(t, "IssueBook")
}

func TestIssueBook_NotAvailable(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	bookRepo.On("GetByID", ctx, int64(1)).Return(unavailableBook(1), nil).Once()

	err := svc.IssueBook(ctx, 1, 10)

	assert.EqualError(t, err, "book is currently not available")
	borrowRepo.AssertNotCalled(t, "IssueBook")
}

func TestIssueBook_HasOverdue(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	bookRepo.On("GetByID", ctx, int64(1)).Return(availableBook(1), nil).Once()
	borrowRepo.On("HasOverdueBorrowing", ctx, int64(10)).Return(true, nil).Once()

	err := svc.IssueBook(ctx, 1, 10)

	assert.EqualError(t, err, "you have overdue books; return them before borrowing more")
	borrowRepo.AssertNotCalled(t, "IssueBook")
}

func TestIssueBook_BorrowLimitReached(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	bookRepo.On("GetByID", ctx, int64(1)).Return(availableBook(1), nil).Once()
	borrowRepo.On("HasOverdueBorrowing", ctx, int64(10)).Return(false, nil).Once()
	borrowRepo.On("CountActiveBorrowings", ctx, int64(10)).Return(maxActiveBorrows, nil).Once()

	err := svc.IssueBook(ctx, 1, 10)

	assert.EqualError(t, err, "borrow limit reached (max 5 books)")
	borrowRepo.AssertNotCalled(t, "IssueBook")
}

func TestIssueBook_AlreadyBorrowed(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	bookRepo.On("GetByID", ctx, int64(1)).Return(availableBook(1), nil).Once()
	borrowRepo.On("HasOverdueBorrowing", ctx, int64(10)).Return(false, nil).Once()
	borrowRepo.On("CountActiveBorrowings", ctx, int64(10)).Return(1, nil).Once()
	borrowRepo.On("HasActiveBorrowing", ctx, int64(1), int64(10)).Return(true, nil).Once()

	err := svc.IssueBook(ctx, 1, 10)

	assert.EqualError(t, err, "you already have an active borrowing for this book")
	borrowRepo.AssertNotCalled(t, "IssueBook")
}

func TestIssueBook_CheckOverdueError(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	bookRepo.On("GetByID", ctx, int64(1)).Return(availableBook(1), nil).Once()
	borrowRepo.On("HasOverdueBorrowing", ctx, int64(10)).Return(false, errors.New("db error")).Once()

	err := svc.IssueBook(ctx, 1, 10)

	assert.EqualError(t, err, "failed to check overdue status")
}

func TestReturnBook_Success(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	borrowRepo.On("ReturnBook", ctx, int64(1), int64(10)).Return(nil).Once()

	err := svc.ReturnBook(ctx, 1, 10)

	assert.NoError(t, err)
	borrowRepo.AssertExpectations(t)
}

func TestReturnBook_Error(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	borrowRepo.On("ReturnBook", ctx, int64(1), int64(10)).
		Return(errors.New("no active borrowing found for this book and user")).Once()

	err := svc.ReturnBook(ctx, 1, 10)

	assert.EqualError(t, err, "no active borrowing found for this book and user")
}

func TestGetMyBorrowings_Success(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	expected := []models.Borrowing{
		{ID: 1, BookID: 1, UserID: 10, Status: "borrowed"},
	}
	borrowRepo.On("GetMyBorrowings", ctx, int64(10)).Return(expected, nil).Once()

	result, err := svc.GetMyBorrowings(ctx, 10)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	borrowRepo.AssertExpectations(t)
}

func TestGetMyBorrowings_Error(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	borrowRepo.On("GetMyBorrowings", ctx, int64(10)).
		Return(nil, errors.New("db error")).Once()

	result, err := svc.GetMyBorrowings(ctx, 10)

	assert.Nil(t, result)
	assert.EqualError(t, err, "db error")
}

func TestGetOverdueBorrowings_Success(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	expected := []models.Borrowing{
		{ID: 5, BookID: 2, UserID: 10, Status: "borrowed", FineAmount: 20},
	}
	borrowRepo.On("GetOverdueBorrowings", ctx).Return(expected, nil).Once()

	result, err := svc.GetOverdueBorrowings(ctx)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	borrowRepo.AssertExpectations(t)
}

func TestGetOverdueBorrowings_Error(t *testing.T) {
	borrowRepo := new(MockBorrowingRepository)
	bookRepo := new(MockBookRepoForBorrowing)
	svc := NewBorrowingService(borrowRepo, bookRepo)
	ctx := context.Background()

	borrowRepo.On("GetOverdueBorrowings", ctx).
		Return(nil, errors.New("db error")).Once()

	result, err := svc.GetOverdueBorrowings(ctx)

	assert.Nil(t, result)
	assert.EqualError(t, err, "db error")
}
