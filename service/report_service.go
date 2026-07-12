package service

import (
	"context"
	"database/sql"

	"book-management/models"
)

type ReportService struct {
	db *sql.DB 
}

func NewReportService(db *sql.DB) *ReportService {
	return &ReportService{db: db}
}

func (s *ReportService) GetDashboardStats(ctx context.Context) (models.Report, error) {
	return models.Report{}, nil
}

func (s *ReportService) GetOverdueBooks(ctx context.Context) ([]models.Borrowing, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT b.id, b.book_id, bk.title, b.user_id, u.email,
			b.issue_date, b.due_date, b.return_date, b.status, b.fine_amount
		FROM borrowings b
		JOIN books bk ON bk.id = b.book_id
		JOIN users u ON u.id = b.user_id
		WHERE b.due_date < NOW() AND b.status = 'borrowed'
		ORDER BY b.due_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overdue []models.Borrowing
	for rows.Next() {
		var b models.Borrowing
		if err := rows.Scan(&b.ID, &b.BookID, &b.BookTitle, &b.UserID, &b.UserEmail,
			&b.IssueDate, &b.DueDate, &b.ReturnDate, &b.Status, &b.FineAmount); err != nil {
			return nil, err
		}
		overdue = append(overdue, b)
	}
	return overdue, rows.Err()
}