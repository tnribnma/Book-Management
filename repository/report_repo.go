package repository

import (
	"context"
	"database/sql"
	"fmt"

	"book-management/models"
)

type ReportRepository interface {
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
	GetPopularBooks(ctx context.Context, limit int) ([]models.Book, error)
	GetActiveBorrowingsCount(ctx context.Context) (int, error)
	GetOverdueCount(ctx context.Context) (int, error)
	GetTotalBooksCount(ctx context.Context) (int, error)
	GetTotalUsersCount(ctx context.Context) (int, error)
}

type reportRepo struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) ReportRepository {
	return &reportRepo{db: db}
}

func (r *reportRepo) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&stats.TotalBooks)
	if err != nil {
		return nil, fmt.Errorf("failed to get total books: %w", err)
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books WHERE status = 'available'").Scan(&stats.AvailableBooks)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books WHERE status = 'borrowed'").Scan(&stats.BorrowedBooks)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM borrowings WHERE status = 'borrowed'").Scan(&stats.ActiveBorrowings)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM borrowings 
		WHERE status = 'borrowed' AND due_date < NOW()`).Scan(&stats.OverdueBorrowings)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reservations WHERE status = 'pending'").Scan(&stats.PendingReservations)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *reportRepo) GetPopularBooks(ctx context.Context, limit int) ([]models.Book, error) {
	query := `
		SELECT b.id, b.title, b.author, COUNT(br.id) as borrow_count
		FROM books b
		LEFT JOIN borrowings br ON b.id = br.book_id
		GROUP BY b.id, b.title, b.author
		ORDER BY borrow_count DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular books: %w", err)
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var borrowCount int
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &borrowCount); err != nil {
			return nil, err
		}
		b.Quantity = borrowCount 
		books = append(books, b)
	}

	return books, nil
}

func (r *reportRepo) GetActiveBorrowingsCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM borrowings WHERE status = 'borrowed'").Scan(&count)
	return count, err
}

func (r *reportRepo) GetOverdueCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM borrowings 
		WHERE status = 'borrowed' AND due_date < NOW()`).Scan(&count)
	return count, err
}

func (r *reportRepo) GetTotalBooksCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&count)
	return count, err
}

func (r *reportRepo) GetTotalUsersCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}