package service

import (
	"sort"
	"strings"
	"book-management/models"
)

type SortStrategy interface {
	Sort(books []models.Book)
}

type dateSorter struct{}

func (s *dateSorter) Sort(books []models.Book) {
	sort.Slice(books, func(i, j int) bool {
		return books[i].CreatedAt.After(books[j].CreatedAt)
	})
}

type titleSorter struct{}

func (s *titleSorter) Sort(books []models.Book) {
	sort.Slice(books, func(i, j int) bool {
		return strings.ToLower(books[i].Title) < strings.ToLower(books[j].Title)
	})
}

type availabilitySorter struct{}

func (s *availabilitySorter) Sort(books []models.Book) {
	sort.Slice(books, func(i, j int) bool {
		return books[i].AvailableCopies > books[j].AvailableCopies
	})
}

type quantitySorter struct{}

func (s *quantitySorter) Sort(books []models.Book) {
	sort.Slice(books, func(i, j int) bool {
		return books[i].Quantity > books[j].Quantity
	})
}

type authorSorter struct{}

func (s *authorSorter) Sort(books []models.Book) {
	sort.Slice(books, func(i, j int) bool {
		return strings.ToLower(books[i].Author) < strings.ToLower(books[j].Author)
	})
}

func GetSorter(sortType string) SortStrategy {
	switch strings.ToLower(sortType) {
	case "date":
		return &dateSorter{}
	case "title":
		return &titleSorter{}
	case "availability":
		return &availabilitySorter{}
	case "quantity":
		return &quantitySorter{}
	case "author":
		return &authorSorter{}
	default:
		return &dateSorter{}
	}
}