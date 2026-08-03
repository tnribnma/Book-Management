package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"book-management/models"
	"book-management/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBookService struct {
	mock.Mock
}

func (m *MockBookService) CreateBook(ctx context.Context, req models.BookRequest, userID int64) (*models.Book, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) GetBook(ctx context.Context, id int64) (*models.Book, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) ListBooks(ctx context.Context, filter models.BookFilter) ([]models.Book, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

func (m *MockBookService) UpdateBook(ctx context.Context, id int64, req models.BookRequest) (*models.Book, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Book), args.Error(1)
}

func (m *MockBookService) DeleteBook(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBookService) UpdateAvailability(ctx context.Context, bookID int64, delta int) error {
	args := m.Called(ctx, bookID, delta)
	return args.Error(0)
}

func (m *MockBookService) SearchBooks(ctx context.Context, query, searchType, sortType string) ([]models.Book, error) {
	args := m.Called(ctx, query, searchType, sortType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

func (m *MockBookService) ListBooksSorted(ctx context.Context, filter models.BookFilter, sortType string) ([]models.Book, error) {
	args := m.Called(ctx, filter, sortType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Book), args.Error(1)
}

func (m *MockBookService) AddBookLink(ctx context.Context, bookID int64, url, bookType string) error {
	args := m.Called(ctx, bookID, url, bookType)
	return args.Error(0)
}

var _ service.BookService = (*MockBookService)(nil)

func int64Ptr(v int64) *int64 { return &v }

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
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	reqBody := sampleBookRequest()
	raw, _ := json.Marshal(reqBody)
	book := sampleBook()
	mockSvc.On("CreateBook", mock.Anything, reqBody, int64(0)).Return(book, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateBook(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp models.Book
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, book.Title, resp.Title)
	assert.Equal(t, book.ID, resp.ID)
	mockSvc.AssertExpectations(t)
}

func TestCreateBook_InvalidJSON(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader([]byte(`{bad`)))
	rr := httptest.NewRecorder()

	h.CreateBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid JSON", resp["error"])
	mockSvc.AssertNotCalled(t, "CreateBook")
}

func TestCreateBook_ValidationFailed(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)
	body := `{"title":"","author":"","quantity":0}`
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader([]byte(body)))
	rr := httptest.NewRecorder()

	h.CreateBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Contains(t, resp["error"], "Validation failed")
	mockSvc.AssertNotCalled(t, "CreateBook")
}

func TestCreateBook_ServiceError(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	reqBody := sampleBookRequest()
	raw, _ := json.Marshal(reqBody)

	mockSvc.On("CreateBook", mock.Anything, reqBody, int64(0)).
		Return(nil, errors.New("quantity must be greater than 0")).Once()

	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(raw))
	rr := httptest.NewRecorder()

	h.CreateBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "quantity must be greater than 0", resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestListBooks_Success(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	books := []models.Book{*sampleBook()}
	expectedFilter := models.BookFilter{
		Search:   "go",
		Author:   "Donovan",
		Status:   "available",
		Category: 1,
	}

	mockSvc.On("ListBooksSorted", mock.Anything, expectedFilter, "title").
		Return(books, nil).Once()

	req := httptest.NewRequest(http.MethodGet,
		"/books?search=go&author=Donovan&status=available&category_id=1&sort=title", nil)
	rr := httptest.NewRecorder()

	h.ListBooks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []models.Book
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "The Go Programming Language", resp[0].Title)
	mockSvc.AssertExpectations(t)
}

func TestListBooks_DefaultSort(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("ListBooksSorted", mock.Anything, models.BookFilter{}, "date").
		Return([]models.Book{}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rr := httptest.NewRecorder()

	h.ListBooks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestListBooks_ServiceError(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("ListBooksSorted", mock.Anything, mock.Anything, "date").
		Return(nil, errors.New("db error")).Once()

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rr := httptest.NewRecorder()

	h.ListBooks(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Failed to fetch books", resp["error"])
}

func TestGetBook_Success(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)
	book := sampleBook()

	mockSvc.On("GetBook", mock.Anything, int64(1)).Return(book, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/books/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.GetBook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Book
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, book.Title, resp.Title)
	mockSvc.AssertExpectations(t)
}

func TestGetBook_InvalidID(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/books/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()

	h.GetBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid book ID", resp["error"])
	mockSvc.AssertNotCalled(t, "GetBook")
}

func TestGetBook_NotFound(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("GetBook", mock.Anything, int64(99)).
		Return(nil, errors.New("book not found")).Once()

	req := httptest.NewRequest(http.MethodGet, "/books/99", nil)
	req.SetPathValue("id", "99")
	rr := httptest.NewRecorder()

	h.GetBook(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Book not found", resp["error"])
}

func TestUpdateBook_Success(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	reqBody := sampleBookRequest()
	reqBody.Title = "Updated Title"
	raw, _ := json.Marshal(reqBody)
	updated := sampleBook()
	updated.Title = "Updated Title"

	mockSvc.On("UpdateBook", mock.Anything, int64(1), reqBody).Return(updated, nil).Once()

	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(raw))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.UpdateBook(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Book
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", resp.Title)
	mockSvc.AssertExpectations(t)
}

func TestUpdateBook_InvalidID(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/books/xyz", bytes.NewReader([]byte(`{}`)))
	req.SetPathValue("id", "xyz")
	rr := httptest.NewRecorder()

	h.UpdateBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "UpdateBook")
}

func TestUpdateBook_InvalidJSON(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader([]byte(`{`)))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.UpdateBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid request body", resp["error"])
}

func TestUpdateBook_ServiceError(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	reqBody := sampleBookRequest()
	raw, _ := json.Marshal(reqBody)

	mockSvc.On("UpdateBook", mock.Anything, int64(1), reqBody).
		Return(nil, errors.New("book not found")).Once()

	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(raw))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.UpdateBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteBook_Success(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("DeleteBook", mock.Anything, int64(1)).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.DeleteBook(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestDeleteBook_InvalidID(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/books/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()

	h.DeleteBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "DeleteBook")
}

func TestDeleteBook_ServiceError(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("DeleteBook", mock.Anything, int64(1)).
		Return(errors.New("cannot delete book that is currently borrowed")).Once()

	req := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.DeleteBook(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "cannot delete book that is currently borrowed", resp["error"])
}

func TestSearchBooks_Success(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	books := []models.Book{*sampleBook()}
	mockSvc.On("SearchBooks", mock.Anything, "go", "title", "date").
		Return(books, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/books/search?query=go&search_type=title&sort=date", nil)
	rr := httptest.NewRecorder()

	h.SearchBooks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []models.Book
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp, 1)
	mockSvc.AssertExpectations(t)
}

func TestSearchBooks_MissingQuery(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/books/search", nil)
	rr := httptest.NewRecorder()

	h.SearchBooks(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "query parameter required", resp["error"])
	mockSvc.AssertNotCalled(t, "SearchBooks")
}

func TestSearchBooks_Defaults(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("SearchBooks", mock.Anything, "go", "title", "date").
		Return([]models.Book{}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/books/search?query=go", nil)
	rr := httptest.NewRecorder()

	h.SearchBooks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestSearchBooks_ServiceError(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("SearchBooks", mock.Anything, "go", "title", "date").
		Return(nil, errors.New("db down")).Once()

	req := httptest.NewRequest(http.MethodGet, "/books/search?query=go", nil)
	rr := httptest.NewRecorder()

	h.SearchBooks(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Contains(t, resp["error"], "Search failed")
}

func TestSearchBooksLegacy_WithQuery(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	books := []models.Book{*sampleBook()}
	mockSvc.On("SearchBooks", mock.Anything, "go", "title", "date").
		Return(books, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/books?q=go", nil)
	rr := httptest.NewRecorder()

	h.SearchBooksLegacy(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestSearchBooksLegacy_NoQuery_FallsBackToList(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	mockSvc.On("ListBooksSorted", mock.Anything, models.BookFilter{}, "date").
		Return([]models.Book{}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rr := httptest.NewRecorder()

	h.SearchBooksLegacy(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestAddBookLink_Success(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	body := map[string]string{"url": "https://example.com/book.pdf", "type": "pdf"}
	raw, _ := json.Marshal(body)
	book := sampleBook()
	book.BookURL = body["url"]
	book.BookType = body["type"]

	mockSvc.On("AddBookLink", mock.Anything, int64(1), body["url"], body["type"]).
		Return(nil).Once()
	mockSvc.On("GetBook", mock.Anything, int64(1)).Return(book, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/books/1/link", bytes.NewReader(raw))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.AddBookLink(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.Book
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, body["url"], resp.BookURL)
	mockSvc.AssertExpectations(t)
}

func TestAddBookLink_InvalidID(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/books/abc/link", bytes.NewReader([]byte(`{}`)))
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()

	h.AddBookLink(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "AddBookLink")
}

func TestAddBookLink_InvalidJSON(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/books/1/link", bytes.NewReader([]byte(`{`)))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.AddBookLink(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Invalid JSON", resp["error"])
}

func TestAddBookLink_ValidationFailed(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	body := `{"url":"","type":"docx"}`
	req := httptest.NewRequest(http.MethodPost, "/books/1/link", bytes.NewReader([]byte(body)))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.AddBookLink(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Contains(t, resp["error"], "Validation failed")
	mockSvc.AssertNotCalled(t, "AddBookLink")
}

func TestAddBookLink_ServiceError(t *testing.T) {
	mockSvc := new(MockBookService)
	h := NewBookHandler(mockSvc)

	body := map[string]string{"url": "https://example.com/book.pdf", "type": "pdf"}
	raw, _ := json.Marshal(body)

	mockSvc.On("AddBookLink", mock.Anything, int64(1), body["url"], body["type"]).
		Return(errors.New("book not found")).Once()

	req := httptest.NewRequest(http.MethodPost, "/books/1/link", bytes.NewReader(raw))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.AddBookLink(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "book not found", resp["error"])
	mockSvc.AssertExpectations(t)
}
