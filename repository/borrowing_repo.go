package repository

import (
	"context"
	"database/sql"
	"book-management/models"
)

type BorrowingRepository struct {
	db *sql.DB
}

func NewBorrowingRepository(db *sql.DB) *BorrowingRepository {
	return &BorrowingRepository{db: db}
}

func (r *BorrowingRepository) Create(ctx context.Context, borrowing models.Borrowing) (models.Borrowing, error) {
	var b models.Borrowing
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO borrowings (book_id, user_id, due_date, status)
		VALUES ($1, $2, $3, 'borrowed') 
		RETURNING id, book_id, user_id, issue_date, due_date, return_date, status, fine_amount`,
		borrowing.BookID, borrowing.UserID, borrowing.DueDate).Scan(
		&b.ID, &b.BookID, &b.UserID, &b.IssueDate, &b.DueDate,
		&b.ReturnDate, &b.Status, &b.FineAmount)
	return b, err
}

func (r *BorrowingRepository) GetByID(ctx context.Context, id int64) (models.Borrowing, error) {
	var b models.Borrowing
	err := r.db.QueryRowContext(ctx, `
		SELECT id, book_id, user_id, issue_date, due_date, return_date, status, fine_amount 
		FROM borrowings 
		WHERE id = $1`, id).Scan(
		&b.ID, &b.BookID, &b.UserID, &b.IssueDate, &b.DueDate,
		&b.ReturnDate, &b.Status, &b.FineAmount)
	return b, err
}

func (r *BorrowingRepository) ReturnBook(ctx context.Context, borrowingID int64, fine float64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE borrowings 
		SET return_date = NOW(), 
		    status = 'returned', 
		    fine_amount = $1 
		WHERE id = $2`, fine, borrowingID)
	return err
}

func (r *BorrowingRepository) GetUserBorrowings(ctx context.Context, userID int64) ([]models.Borrowing, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, book_id, user_id, issue_date, due_date, return_date, status, fine_amount 
		FROM borrowings 
		WHERE user_id = $1 
		ORDER BY issue_date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var borrowings []models.Borrowing
	for rows.Next() {
		var b models.Borrowing
		err := rows.Scan(
			&b.ID, &b.BookID, &b.UserID, &b.IssueDate, &b.DueDate,
			&b.ReturnDate, &b.Status, &b.FineAmount)
		if err != nil {
			return nil, err
		}
		borrowings = append(borrowings, b)
	}
	return borrowings, nil
}

func (r *BorrowingRepository) GetOverdue(ctx context.Context) ([]models.Borrowing, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, book_id, user_id, issue_date, due_date, return_date, status, fine_amount 
		FROM borrowings 
		WHERE status = 'borrowed' 
		  AND due_date < NOW() 
		ORDER BY due_date ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var borrowings []models.Borrowing
	for rows.Next() {
		var b models.Borrowing
		err := rows.Scan(
			&b.ID, &b.BookID, &b.UserID, &b.IssueDate, &b.DueDate,
			&b.ReturnDate, &b.Status, &b.FineAmount)
		if err != nil {
			return nil, err
		}
		borrowings = append(borrowings, b)
	}
	return borrowings, nil
}