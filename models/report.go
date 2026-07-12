package models

type Report struct {
	TotalBooks      int     `json:"total_books"`
	AvailableBooks  int     `json:"available_books"`
	BorrowedBooks   int     `json:"borrowed_books"`
	OverdueBooks    int     `json:"overdue_books"`
	TotalFines      float64 `json:"total_fines"`
	MostBorrowed    string  `json:"most_borrowed,omitempty"`
}