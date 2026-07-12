package service

import (
	"context"
	"database/sql"

	"book-management/models"
	"book-management/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(db *sql.DB) *CategoryService {
	return &CategoryService{repo: repository.NewCategoryRepository(db)}
}

func (s *CategoryService) Create(ctx context.Context, req models.Category) (models.Category, error) {
	return s.repo.Create(ctx, req)
}

func (s *CategoryService) List(ctx context.Context) ([]models.Category, error) {
	return s.repo.List(ctx)
}