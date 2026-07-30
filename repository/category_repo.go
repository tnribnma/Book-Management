package repository

import (
	"context"
	"database/sql"
	"fmt"

	"book-management/models"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]models.Category, error)
	GetByID(ctx context.Context, id int64) (*models.Category, error)
	Create(ctx context.Context, category *models.Category) error
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id int64) error
}

type categoryRepo struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) List(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, description, created_at 
		FROM categories 
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, nil
}

func (r *categoryRepo) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	query := `
		SELECT id, name, description, created_at 
		FROM categories 
		WHERE id = $1`

	var category models.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

func (r *categoryRepo) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, category.Name, category.Description).
		Scan(&category.ID, &category.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *categoryRepo) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories 
		SET name = $1, description = $2 
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, category.Name, category.Description, category.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *categoryRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}