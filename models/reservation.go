package models

import "time"

type Reservation struct {
	ID              int64     `json:"id"`
	BookID          int64     `json:"book_id"`
	BookTitle       string    `json:"book_title,omitempty"`
	UserID          int64     `json:"user_id"`
	UserName        string    `json:"user_name,omitempty"`
	ReservationDate time.Time `json:"reservation_date"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}