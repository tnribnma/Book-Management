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

func (r *BookRepository) List(ctx context.Context, userID int64, limit, page int) ([]models.Book, int, error) {
	offset := (page - 1) * limit

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, title, author, published_year, created_at 
		FROM books 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.PublishedYear, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		books = append(books, b)
	}

	var total int
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books WHERE user_id = $1`, userID).Scan(&total)
	return books, total, err
}

func (r *BookRepository) Create(ctx context.Context, userID int64, req models.BookRequest) (*models.Book, error) {
	var book models.Book
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO books (user_id, title, author, published_year)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, author, published_year, created_at`,
		userID, req.Title, req.Author, req.PublishedYear,
	).Scan(&book.ID, &book.UserID, &book.Title, &book.Author, &book.PublishedYear, &book.CreatedAt)

	return &book, err
}

func (r *BookRepository) GetByID(ctx context.Context, userID, id int64) (*models.Book, error) {
	var book models.Book
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, author, published_year, created_at 
		FROM books 
		WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&book.ID, &book.UserID, &book.Title, &book.Author, &book.PublishedYear, &book.CreatedAt)

	return &book, err
}

func (r *BookRepository) Update(ctx context.Context, userID, id int64, req models.BookRequest) (*models.Book, error) {
	var book models.Book
	err := r.db.QueryRowContext(ctx, `
		UPDATE books 
		SET title = $1, author = $2, published_year = $3
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, title, author, published_year, created_at`,
		req.Title, req.Author, req.PublishedYear, id, userID,
	).Scan(&book.ID, &book.UserID, &book.Title, &book.Author, &book.PublishedYear, &book.CreatedAt)

	return &book, err
}

func (r *BookRepository) Delete(ctx context.Context, userID, id int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM books WHERE id = $1 AND user_id = $2`,
		id, userID)
	return err
}