package repository

import (
	"context"
	"database/sql"
	"fmt"
	"book-management/models"
)

type BorrowingRepository interface {
	IssueBook(ctx context.Context, borrowing *models.Borrowing) error
	ReturnBook(ctx context.Context, bookID, userID int64) error
	GetMyBorrowings(ctx context.Context, userID int64) ([]models.Borrowing, error)
	GetOverdueBorrowings(ctx context.Context) ([]models.Borrowing, error)
	HasActiveBorrowing(ctx context.Context, bookID, userID int64) (bool, error)  
	CountActiveBorrowings(ctx context.Context, userID int64) (int, error)
	HasOverdueBorrowing(ctx context.Context, userID int64) (bool, error)
}

type borrowingRepo struct {
	db *sql.DB
}

func (r *borrowingRepo) HasActiveBorrowing(ctx context.Context, bookID, userID int64) (bool, error) {
	query := `
		SELECT COUNT(*) FROM borrowings 
		WHERE book_id = $1 AND user_id = $2 AND status = 'borrowed'`

	var count int
	err := r.db.QueryRowContext(ctx, query, bookID, userID).Scan(&count)
	return count > 0, err
}

func NewBorrowingRepository(db *sql.DB) BorrowingRepository {
	return &borrowingRepo{db: db}
}

func (r *borrowingRepo) IssueBook(ctx context.Context, borrowing *models.Borrowing) error {
	query := `
		INSERT INTO borrowings (book_id, user_id, due_date, status)
		VALUES ($1, $2, $3, 'borrowed')
		RETURNING id, issue_date`

	err := r.db.QueryRowContext(ctx, query, 
		borrowing.BookID, 
		borrowing.UserID, 
		borrowing.DueDate,
	).Scan(&borrowing.ID, &borrowing.IssueDate)

	if err != nil {
		return fmt.Errorf("failed to issue book: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE books 
		SET available_copies = available_copies - 1,
		    status = CASE WHEN available_copies - 1 <= 0 THEN 'borrowed' ELSE status END
		WHERE id = $1`, borrowing.BookID)

	if err != nil {
		return fmt.Errorf("failed to update book availability: %w", err)
	}

	return nil
}

func (r *borrowingRepo) ReturnBook(ctx context.Context, bookID, userID int64) error {
	query := `
		UPDATE borrowings 
		SET return_date = NOW(), 
		    status = 'returned',
		    fine_amount = CASE 
		        WHEN NOW() > due_date THEN 
		            EXTRACT(DAY FROM (NOW() - due_date)) * 10.00 
		        ELSE 0.00 
		    END
		WHERE book_id = $1 AND user_id = $2 AND status = 'borrowed'
		RETURNING id`

	var borrowingID int64
	err := r.db.QueryRowContext(ctx, query, bookID, userID).Scan(&borrowingID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no active borrowing found for this book and user")
	}
	if err != nil {
		return fmt.Errorf("failed to return book: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE books 
		SET available_copies = available_copies + 1,
		    status = 'available'
		WHERE id = $1`, bookID)

	if err != nil {
		return fmt.Errorf("failed to update book availability: %w", err)
	}

	return nil
}

func (r *borrowingRepo) CountActiveBorrowings(ctx context.Context, userID int64) (int, error) {
	query := `
		SELECT COUNT(*) FROM borrowings
		WHERE user_id = $1 AND status IN ('borrowed', 'overdue')`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}
 
func (r *borrowingRepo) HasOverdueBorrowing(ctx context.Context, userID int64) (bool, error) {
	query := `
		SELECT COUNT(*) FROM borrowings
		WHERE user_id = $1
		  AND status = 'borrowed'
		  AND due_date < NOW()`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count > 0, err
}

func (r *borrowingRepo) GetMyBorrowings(ctx context.Context, userID int64) ([]models.Borrowing, error) {
	query := `
		SELECT b.id, b.book_id, bk.title as book_title, b.user_id, b.issue_date, 
		       b.due_date, b.return_date, b.status, b.fine_amount, b.created_at
		FROM borrowings b
		JOIN books bk ON b.book_id = bk.id
		WHERE b.user_id = $1
		ORDER BY b.issue_date DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get borrowings: %w", err)
	}
	defer rows.Close()

	var borrowings []models.Borrowing
	for rows.Next() {
		var br models.Borrowing
		if err := rows.Scan(
			&br.ID, &br.BookID, &br.BookTitle, &br.UserID, &br.IssueDate,
			&br.DueDate, &br.ReturnDate, &br.Status, &br.FineAmount, &br.CreatedAt,
		); err != nil {
			return nil, err
		}
		borrowings = append(borrowings, br)
	}

	return borrowings, nil
}

func (r *borrowingRepo) GetOverdueBorrowings(ctx context.Context) ([]models.Borrowing, error) {
	query := `
		SELECT b.id, b.book_id, bk.title as book_title, b.user_id, u.full_name as user_name,
		       b.issue_date, b.due_date, b.status, b.fine_amount
		FROM borrowings b
		JOIN books bk ON b.book_id = bk.id
		JOIN users u ON b.user_id = u.id
		WHERE b.status = 'borrowed' AND b.due_date < NOW()`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var borrowings []models.Borrowing
	for rows.Next() {
		var br models.Borrowing
		if err := rows.Scan(
			&br.ID, &br.BookID, &br.BookTitle, &br.UserID, &br.UserName,
			&br.IssueDate, &br.DueDate, &br.Status, &br.FineAmount,
		); err != nil {
			return nil, err
		}
		borrowings = append(borrowings, br)
	}

	return borrowings, nil
}