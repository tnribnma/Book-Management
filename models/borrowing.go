package models

import "time"

type Borrowing struct {
	ID          int64      `json:"id"`
	BookID      int64      `json:"book_id"`
	BookTitle   string     `json:"book_title,omitempty"`
	UserID      int64      `json:"user_id"`
	UserEmail   string     `json:"user_email,omitempty"`
	IssueDate   time.Time  `json:"issue_date"`
	DueDate     time.Time  `json:"due_date"`
	ReturnDate  *time.Time `json:"return_date,omitempty"`
	Status      string     `json:"status"`
	FineAmount  float64    `json:"fine_amount"`
}

type BorrowRequest struct {
	BookID  int64 `json:"book_id" validate:"required"`
	UserID  int64 `json:"user_id" validate:"required"` 
	DueDays int   `json:"due_days" validate:"required,min=1,max=60"`
}
type ReturnRequest struct {
	BorrowingID int64 `json:"borrowing_id" validate:"required"`
}