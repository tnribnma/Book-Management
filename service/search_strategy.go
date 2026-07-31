package service

import (
	"context"
	"strings"
	"book-management/models"
	"book-management/repository"
)

type SearchStrategy interface {
	Search(ctx context.Context, query string, bookRepo repository.BookRepository) ([]models.Book, error)
}

type titleSearcher struct{}

func (s *titleSearcher) Search(ctx context.Context, query string, bookRepo repository.BookRepository) ([]models.Book, error) {
	filter := models.BookFilter{Search: query}
	return bookRepo.List(ctx, filter)
}

type authorSearcher struct{}

func (s *authorSearcher) Search(ctx context.Context, query string, bookRepo repository.BookRepository) ([]models.Book, error) {
	filter := models.BookFilter{Author: query}
	return bookRepo.List(ctx, filter)
}

type isbnSearcher struct{}

func (s *isbnSearcher) Search(ctx context.Context, query string, bookRepo repository.BookRepository) ([]models.Book, error) {
	filter := models.BookFilter{Search: query}
	books, err := bookRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	filtered := []models.Book{}
	for _, book := range books {
		if book.ISBN == query {
			filtered = append(filtered, book)
		}
	}
	return filtered, nil
}

type fullTextSearcher struct{}

func (s *fullTextSearcher) Search(ctx context.Context, query string, bookRepo repository.BookRepository) ([]models.Book, error) {
	filter := models.BookFilter{Search: query}
	books, err := bookRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	filtered := []models.Book{}
	queryLower := strings.ToLower(query)

	for _, book := range books {
		if strings.Contains(strings.ToLower(book.Title), queryLower) ||
			strings.Contains(strings.ToLower(book.Author), queryLower) ||
			strings.Contains(book.ISBN, query) ||
			strings.Contains(strings.ToLower(book.Publisher), queryLower) {
			filtered = append(filtered, book)
		}
	}
	return filtered, nil
}

type fuzzySearcher struct{}

func (s *fuzzySearcher) Search(ctx context.Context, query string, bookRepo repository.BookRepository) ([]models.Book, error) {
	filter := models.BookFilter{}
	books, err := bookRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	filtered := []models.Book{}
	queryLower := strings.ToLower(query)

	for _, book := range books {
		if fuzzyMatch(strings.ToLower(book.Title), queryLower) ||
			fuzzyMatch(strings.ToLower(book.Author), queryLower) {
			filtered = append(filtered, book)
		}
	}

	return filtered, nil
}

func fuzzyMatch(s, pattern string) bool {
	if strings.Contains(s, pattern) {
		return true
	}
	if len(s) > 0 && len(pattern) > 0 {
		diff := abs(len(s) - len(pattern))
		return diff <= 2
	}
	return false
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func GetSearcher(searchType string) SearchStrategy {
	switch strings.ToLower(searchType) {
	case "author":
		return &authorSearcher{}
	case "isbn":
		return &isbnSearcher{}
	case "fulltext":
		return &fullTextSearcher{}
	case "fuzzy":
		return &fuzzySearcher{}
	default:
		return &titleSearcher{}
	}
}