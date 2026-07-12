package repository

import (
	"context"
	"database/sql"
	"book-management/models"
)

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book models.Book) (models.Book, error) {
	var created models.Book
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO books (title, author, isbn, category_id, publisher, edition, 
			published_year, quantity, available_copies, shelf, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, $9, 'available')
		RETURNING id, title, author, isbn, category_id, publisher, edition, 
			published_year, quantity, available_copies, shelf, status, created_at`,
		book.Title, book.Author, book.ISBN, book.CategoryID, book.Publisher,
		book.Edition, book.PublishedYear, book.Quantity, book.Shelf).Scan(
		&created.ID, &created.Title, &created.Author, &created.ISBN, &created.CategoryID,
		&created.Publisher, &created.Edition, &created.PublishedYear, &created.Quantity,
		&created.AvailableCopies, &created.Shelf, &created.Status, &created.CreatedAt)
	return created, err
}

func (r *BookRepository) List(ctx context.Context, filter models.BookFilter, limit, offset int) ([]models.Book, int, error) {
	query := `
		SELECT id, title, author, isbn, category_id, publisher, edition, 
		       published_year, quantity, available_copies, shelf, status, created_at 
		FROM books WHERE 1=1`
	args := []interface{}{}

	if filter.Search != "" {
		query += " AND (title ILIKE $1 OR author ILIKE $1)"
		args = append(args, "%"+filter.Search+"%")
	}
	if filter.Category != 0 {
		query += " AND category_id = $2"
		args = append(args, filter.Category)
	}
	if filter.Status != "" {
		query += " AND status = $3"
		args = append(args, filter.Status)
	}

	var total int
	countErr := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books WHERE 1=1").Scan(&total)
	if countErr != nil {
		return nil, 0, countErr
	}

	query += " ORDER BY created_at DESC LIMIT $4 OFFSET $5"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.CategoryID,
			&b.Publisher, &b.Edition, &b.PublishedYear, &b.Quantity,
			&b.AvailableCopies, &b.Shelf, &b.Status, &b.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		books = append(books, b)
	}
	return books, total, nil
}

func (r *BookRepository) GetByID(ctx context.Context, id int64) (models.Book, error) {
	var book models.Book
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, author, isbn, category_id, publisher, edition, 
		       published_year, quantity, available_copies, shelf, status, created_at 
		FROM books WHERE id = $1`, id).Scan(
		&book.ID, &book.Title, &book.Author, &book.ISBN, &book.CategoryID,
		&book.Publisher, &book.Edition, &book.PublishedYear, &book.Quantity,
		&book.AvailableCopies, &book.Shelf, &book.Status, &book.CreatedAt)
	return book, err
}

func (r *BookRepository) UpdateAvailability(ctx context.Context, bookID int64, change int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE books SET available_copies = available_copies + $1 
		WHERE id = $2`, change, bookID)
	return err
}

func (r *BookRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM books WHERE id = $1", id)
	return err
}

func (r *BookRepository) Search(ctx context.Context, query string) ([]models.Book, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, author, isbn, category_id, publisher, edition, 
		       published_year, quantity, available_copies, shelf, status, created_at 
		FROM books 
		WHERE title ILIKE $1 OR author ILIKE $1 
		LIMIT 50`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		err := rows.Scan(
			&b.ID, &b.Title, &b.Author, &b.ISBN, &b.CategoryID,
			&b.Publisher, &b.Edition, &b.PublishedYear, &b.Quantity,
			&b.AvailableCopies, &b.Shelf, &b.Status, &b.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}