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

func (s *BookService) Create(ctx context.Context, userID int64, req models.BookRequest) (*models.Book, error) {
	if req.Title == "" || req.Author == "" {
		return nil, errors.New("title and author are required")
	}
	return s.repo.Create(ctx, userID, req)
}

func (s *BookService) List(ctx context.Context, userID int64, limit, page int) ([]models.Book, int, error) {
	return s.repo.List(ctx, userID, limit, page)
}

func (s *BookService) GetByID(ctx context.Context, userID, id int64) (*models.Book, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *BookService) Update(ctx context.Context, userID, id int64, req models.BookRequest) (*models.Book, error) {
	if req.Title == "" || req.Author == "" {
		return nil, errors.New("title and author are required")
	}
	return s.repo.Update(ctx, userID, id, req)
}

func (s *BookService) Delete(ctx context.Context, userID, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}