package service

import (
	"context"
	"fmt"

	"book-management/models"
	"book-management/repository"
)

type CategoryService interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
	GetCategory(ctx context.Context, id int64) (*models.Category, error)
	CreateCategory(ctx context.Context, name, description string) (*models.Category, error)
	UpdateCategory(ctx context.Context, id int64, name, description string) (*models.Category, error)
	DeleteCategory(ctx context.Context, id int64) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
	bookRepo     repository.BookRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository, bookRepo repository.BookRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		bookRepo:     bookRepo,
	}
}

func (s *categoryService) ListCategories(ctx context.Context) ([]models.Category, error) {
	return s.categoryRepo.List(ctx)
}

func (s *categoryService) GetCategory(ctx context.Context, id int64) (*models.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *categoryService) CreateCategory(ctx context.Context, name, description string) (*models.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	category := &models.Category{
		Name:        name,
		Description: description,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, id int64, name, description string) (*models.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	category := &models.Category{
		ID:          id,
		Name:        name,
		Description: description,
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int64) error {
	count, err := s.bookRepo.CountByCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check usage")
	}
	if count > 0 {
		return fmt.Errorf("cannot delete: category is used by %d book(s)", count)
	}

	return s.categoryRepo.Delete(ctx, id)
}