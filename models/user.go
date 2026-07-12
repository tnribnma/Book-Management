package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name,omitempty"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FullName  string `json:"full_name,omitempty"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserProfile struct {
	User            User         `json:"user"`
	CurrentBorrowed []Borrowing  `json:"current_borrowed,omitempty"`
	BorrowHistory   []Borrowing  `json:"borrow_history,omitempty"`
}