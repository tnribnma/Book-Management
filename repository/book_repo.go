package repository

import (
	"context"
	"database/sql"
	"fmt"

	"book-management/models"
)

type BookRepository interface {
	List(ctx context.Context, filter models.BookFilter) ([]models.Book, error)
	GetByID(ctx context.Context, id int64) (*models.Book, error)
	Create(ctx context.Context, book *models.Book) error
	Update(ctx context.Context, book *models.Book) error
	Delete(ctx context.Context, id int64) error
	UpdateAvailability(ctx context.Context, bookID int64, delta int) error
	CountByCategory(ctx context.Context, categoryID int64) (int, error)
}

type bookRepo struct {
	db *sql.DB
}

func (r *bookRepo) CountByCategory(ctx context.Context, categoryID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books WHERE category_id = $1", categoryID).Scan(&count)
	return count, err
}

func NewBookRepository(db *sql.DB) BookRepository {
	return &bookRepo{db: db}
}

func (r *bookRepo) List(ctx context.Context, filter models.BookFilter) ([]models.Book, error) {
	query := `
	SELECT b.id, b.title, b.author,
	       COALESCE(b.isbn, '') as isbn,
	       b.category_id,
	       COALESCE(c.name, '') as category_name,
	       COALESCE(b.publisher, '') as publisher,
	       COALESCE(b.edition, '') as edition,
	       COALESCE(b.published_year, 0) as published_year,
	       b.quantity, b.available_copies,
	       COALESCE(b.shelf, '') as shelf,
	       COALESCE(b.status, 'available') as status,
	       b.created_at
	FROM books b
	LEFT JOIN categories c ON b.category_id = c.id
	WHERE 1=1`

	var args []interface{}
	argCount := 1

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (b.title I LIKE $%d OR b.author I LIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}
	if filter.Category != 0 {
		query += fmt.Sprintf(" AND b.category_id = $%d", argCount)
		args = append(args, filter.Category)
		argCount++
	}
	if filter.Author != "" {
		query += fmt.Sprintf(" AND b.author ILIKE $%d", argCount)
		args = append(args, "%"+filter.Author+"%")
		argCount++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND b.status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	query += " ORDER BY b.created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list books: %w", err)
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.CategoryID, &b.CategoryName,
			&b.Publisher, &b.Edition, &b.PublishedYear, &b.Quantity,
			&b.AvailableCopies, &b.Shelf, &b.Status, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}
		books = append(books, b)
	}

	return books, nil
}

func (r *bookRepo) GetByID(ctx context.Context, id int64) (*models.Book, error) {
	query := `
		SELECT b.id, b.title, b.author, b.isbn, b.category_id, COALESCE(c.name, '') as category_name,
		       b.publisher, b.edition, b.published_year, b.quantity, b.available_copies,
		       b.shelf, b.status, b.created_at 
		FROM books b 
		LEFT JOIN categories c ON b.category_id = c.id 
		WHERE b.id = $1`

	var book models.Book
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&book.ID, &book.Title, &book.Author, &book.ISBN, &book.CategoryID, &book.CategoryName,
		&book.Publisher, &book.Edition, &book.PublishedYear, &book.Quantity,
		&book.AvailableCopies, &book.Shelf, &book.Status, &book.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("book not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	return &book, nil
}

func (r *bookRepo) Create(ctx context.Context, book *models.Book) error {
	query := `
		INSERT INTO books (title, author, isbn, category_id, publisher, edition, 
			published_year, quantity, available_copies, shelf, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		book.Title, book.Author, book.ISBN, book.CategoryID,
		book.Publisher, book.Edition, book.PublishedYear,
		book.Quantity, book.AvailableCopies, book.Shelf, book.Status,
	).Scan(&book.ID, &book.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create book: %w", err)
	}
	return nil
}

func (r *bookRepo) Update(ctx context.Context, book *models.Book) error {
	query := `
		UPDATE books SET title=$1, author=$2, isbn=$3, category_id=$4, publisher=$5,
			edition=$6, published_year=$7, quantity=$8, shelf=$9, status=$10
		WHERE id = $11`

	_, err := r.db.ExecContext(ctx, query,
		book.Title, book.Author, book.ISBN, book.CategoryID,
		book.Publisher, book.Edition, book.PublishedYear,
		book.Quantity, book.Shelf, book.Status, book.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update book: %w", err)
	}
	return nil
}

func (r *bookRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM books WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}
	return nil
}

func (r *bookRepo) UpdateAvailability(ctx context.Context, bookID int64, delta int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE books 
		SET available_copies = available_copies + $1 
		WHERE id = $2`, delta, bookID)
	if err != nil {
		return fmt.Errorf("failed to update availability: %w", err)
	}
	return nil
}