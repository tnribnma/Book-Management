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

func int64Ptr(v int64) *int64 { return &v }

type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) List(ctx context.Context, filter models.BookFilter) ([]models.Book, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

func (m *MockBookRepository) GetByID(ctx context.Context, id int64) (*models.Book, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookRepository) Create(ctx context.Context, book *models.Book) error {
	args := m.Called(ctx, book)
	return args.Error(0)
}

func (m *MockBookRepository) Update(ctx context.Context, book *models.Book) error {
	args := m.Called(ctx, book)
	return args.Error(0)
}

func (m *MockBookRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookRepository) UpdateAvailability(ctx context.Context, bookID int64, delta int) error {
	args := m.Called(ctx, bookID, delta)
	return args.Error(0)
}

func (m *MockBookRepository) CountByCategory(ctx context.Context, categoryID int64) (int, error) {
	args := m.Called(ctx, categoryID)
	return args.Int(0), args.Error(1)
}

func (m *MockBookRepository) UpdateBookLink(ctx context.Context, bookID int64, url, bookType string) error {
	args := m.Called(ctx, bookID, url, bookType)
	return args.Error(0)
}

var _ repository.BookRepository = (*MockBookRepository)(nil)

func sampleBookRequest() models.BookRequest {
	return models.BookRequest{
		Title:         "The Go Programming Language",
		Author:        "Alan Donovan",
		ISBN:          "9780134190440",
		CategoryID:    int64Ptr(1),
		Publisher:     "Addison-Wesley",
		Edition:       "1st",
		PublishedYear: 2015,
		Quantity:      5,
		Shelf:         "A-12",
		BookURL:       "https://example.com/go-book",
		BookType:      "pdf",
	}
}

func sampleBook() *models.Book {
	return &models.Book{
		ID:              1,
		Title:           "The Go Programming Language",
		Author:          "Alan Donovan",
		ISBN:            "9780134190440",
		CategoryID:      int64Ptr(1),
		CategoryName:    "Programming",
		Publisher:       "Addison-Wesley",
		Edition:         "1st",
		PublishedYear:   2015,
		Quantity:        5,
		AvailableCopies: 5,
		Shelf:           "A-12",
		Status:          "available",
		BookURL:         "https://example.com/go-book",
		BookType:        "pdf",
		CreatedAt:       time.Now(),
	}
}

func TestCreateBook_Success(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	req := sampleBookRequest()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(b *models.Book) bool {
		return b.Title == req.Title &&
			b.Author == req.Author &&
			b.Quantity == req.Quantity &&
			b.AvailableCopies == req.Quantity &&
			b.Status == "available"
	})).Return(nil).Once()

	book, err := svc.CreateBook(ctx, req, 42)

	require.NoError(t, err)
	assert.NotNil(t, book)
	assert.Equal(t, req.Title, book.Title)
	assert.Equal(t, req.Quantity, book.AvailableCopies)
	assert.Equal(t, "available", book.Status)
	mockRepo.AssertExpectations(t)
}

func TestCreateBook_InvalidQuantity(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)

	req := sampleBookRequest()
	req.Quantity = 0

	book, err := svc.CreateBook(context.Background(), req, 1)

	assert.Nil(t, book)
	assert.EqualError(t, err, "quantity must be greater than 0")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateBook_RepoError(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	req := sampleBookRequest()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Book")).
		Return(errors.New("db error")).Once()

	book, err := svc.CreateBook(ctx, req, 1)

	assert.Nil(t, book)
	assert.ErrorContains(t, err, "failed to create book")
	mockRepo.AssertExpectations(t)
}

func TestGetBook_Success(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	expected := sampleBook()

	mockRepo.On("GetByID", ctx, int64(1)).Return(expected, nil).Once()

	book, err := svc.GetBook(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expected, book)
	mockRepo.AssertExpectations(t)
}

func TestGetBook_NotFound(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(99)).
		Return(nil, errors.New("book not found")).Once()

	book, err := svc.GetBook(ctx, 99)

	assert.Nil(t, book)
	assert.EqualError(t, err, "book not found")
	mockRepo.AssertExpectations(t)
}

func TestListBooks(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	filter := models.BookFilter{Category: 1, Status: "available"}
	expected := []models.Book{*sampleBook()}

	mockRepo.On("List", ctx, filter).Return(expected, nil).Once()

	books, err := svc.ListBooks(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, expected, books)
	mockRepo.AssertExpectations(t)
}

func TestUpdateBook_Success(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	existing := sampleBook()
	req := sampleBookRequest()
	req.Title = "Updated Title"
	req.Quantity = 10

	mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
	mockRepo.On("Update", ctx, mock.MatchedBy(func(b *models.Book) bool {
		return b.Title == "Updated Title" && b.Quantity == 10
	})).Return(nil).Once()

	book, err := svc.UpdateBook(ctx, 1, req)

	require.NoError(t, err)
	assert.Equal(t, "Updated Title", book.Title)
	assert.Equal(t, 10, book.Quantity)
	mockRepo.AssertExpectations(t)
}

func TestUpdateBook_NotFound(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(99)).
		Return(nil, errors.New("book not found")).Once()

	book, err := svc.UpdateBook(ctx, 99, sampleBookRequest())

	assert.Nil(t, book)
	assert.EqualError(t, err, "book not found")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestDeleteBook_Success(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	book := sampleBook()

	mockRepo.On("GetByID", ctx, int64(1)).Return(book, nil).Once()
	mockRepo.On("Delete", ctx, int64(1)).Return(nil).Once()

	err := svc.DeleteBook(ctx, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteBook_CurrentlyBorrowed(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	book := sampleBook()
	book.AvailableCopies = 3

	mockRepo.On("GetByID", ctx, int64(1)).Return(book, nil).Once()

	err := svc.DeleteBook(ctx, 1)

	assert.EqualError(t, err, "cannot delete book that is currently borrowed")
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestDeleteBook_NotFound(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(99)).
		Return(nil, errors.New("book not found")).Once()

	err := svc.DeleteBook(ctx, 99)

	assert.EqualError(t, err, "book not found")
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestUpdateAvailability(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateAvailability", ctx, int64(1), -1).Return(nil).Once()

	err := svc.UpdateAvailability(ctx, 1, -1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddBookLink_Success(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	book := sampleBook()

	mockRepo.On("GetByID", ctx, int64(1)).Return(book, nil).Once()
	mockRepo.On("UpdateBookLink", ctx, int64(1), "https://new-url.com", "epub").
		Return(nil).Once()

	err := svc.AddBookLink(ctx, 1, "https://new-url.com", "epub")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddBookLink_EmptyURL(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)

	err := svc.AddBookLink(context.Background(), 1, "", "pdf")

	assert.EqualError(t, err, "book URL cannot be empty")
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestAddBookLink_InvalidType(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)

	err := svc.AddBookLink(context.Background(), 1, "https://x.com", "docx")

	assert.EqualError(t, err, "invalid book type, must be: link, pdf, or epub")
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestAddBookLink_BookNotFound(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(99)).
		Return(nil, errors.New("book not found")).Once()

	err := svc.AddBookLink(ctx, 99, "https://x.com", "pdf")

	assert.EqualError(t, err, "book not found")
	mockRepo.AssertNotCalled(t, "UpdateBookLink")
}

func TestSearchBooks_EmptyQuery(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)

	books, err := svc.SearchBooks(context.Background(), "", "title", "title")

	require.NoError(t, err)
	assert.Empty(t, books)
}

func TestListBooksSorted(t *testing.T) {
	mockRepo := new(MockBookRepository)
	svc := NewBookService(mockRepo)
	ctx := context.Background()
	filter := models.BookFilter{}
	books := []models.Book{
		{ID: 2, Title: "Zebra", CreatedAt: time.Now()},
		{ID: 1, Title: "Apple", CreatedAt: time.Now().Add(-time.Hour)},
	}

	mockRepo.On("List", ctx, filter).Return(books, nil).Once()

	result, err := svc.ListBooksSorted(ctx, filter, "title")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Apple", result[0].Title)
	assert.Equal(t, "Zebra", result[1].Title)
	mockRepo.AssertExpectations(t)
}
