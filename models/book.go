package models

import "time"

type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	ISBN            string    `json:"isbn,omitempty"`
	CategoryID      *int64    `json:"category_id,omitempty"`
	CategoryName    string    `json:"category_name,omitempty"`
	Publisher       string    `json:"publisher,omitempty"`
	Edition         string    `json:"edition,omitempty"`
	PublishedYear   int       `json:"published_year,omitempty"`
	Quantity        int       `json:"quantity"`
	AvailableCopies int       `json:"available_copies"`
	Shelf           string    `json:"shelf,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type BookRequest struct {
	Title         string  `json:"title" validate:"required,min=1,max=255"`
	Author        string  `json:"author" validate:"required,min=1,max=255"`
	ISBN          string  `json:"isbn,omitempty" validate:"omitempty,isbn"`
	CategoryID    *int64  `json:"category_id,omitempty"`
	Publisher     string  `json:"publisher,omitempty" validate:"omitempty,max=100"`
	Edition       string  `json:"edition,omitempty" validate:"omitempty,max=50"`
	PublishedYear int     `json:"published_year,omitempty" validate:"omitempty,min=1000,max=2100"`
	Quantity      int     `json:"quantity" validate:"required,min=1"`
	Shelf         string  `json:"shelf,omitempty" validate:"omitempty,shelf"`
}

type BookFilter struct {
	Search   string `json:"search,omitempty"`
	Category int64  `json:"category,omitempty"`
	Author   string `json:"author,omitempty"`
	Status   string `json:"status,omitempty"`
}