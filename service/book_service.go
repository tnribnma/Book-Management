package service

import (
	"context"
	"database/sql"
	"errors"

	"book-management/models"
	"book-management/repository"
)

type BookService struct {
	repo *repository.BookRepository
}

func NewBookService(db *sql.DB) *BookService {
	return &BookService{repo: repository.NewBookRepository(db)}
}

func (s *BookService) Create(ctx context.Context, req models.BookRequest) (models.Book, error) {
	if req.Quantity < 1 {
		return models.Book{}, errors.New("quantity must be at least 1")
	}

	book := models.Book{
		Title:          req.Title,
		Author:         req.Author,
		ISBN:           req.ISBN,
		CategoryID:     req.CategoryID,
		Publisher:      req.Publisher,
		Edition:        req.Edition,
		PublishedYear:  req.PublishedYear,
		Quantity:       req.Quantity,
		AvailableCopies: req.Quantity,
		Shelf:          req.Shelf,
		Status:         "available",
	}

	return s.repo.Create(ctx, book)
}

func (s *BookService) List(ctx context.Context, filter models.BookFilter, limit, offset int) ([]models.Book, int, error) {
	return s.repo.List(ctx, filter, limit, offset)
}

func (s *BookService) GetByID(ctx context.Context, id int64) (models.Book, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BookService) Update(ctx context.Context, id int64, req models.BookRequest) (models.Book, error) {
	return models.Book{}, errors.New("update functionality not fully implemented yet")
}

func (s *BookService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *BookService) Search(ctx context.Context, query string) ([]models.Book, error) {
	return s.repo.Search(ctx, query)
}