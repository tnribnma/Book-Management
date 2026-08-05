package repository

import (
	"context"
	"database/sql"
	"testing"

	"book-management/models"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func int64Ptr(v int64) *int64 {
	return &v
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := "postgres://postgres:1234@localhost:5432/books_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	_, err = db.Exec(`
		TRUNCATE books, categories RESTART IDENTITY CASCADE;
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO categories (id, name) VALUES (1, 'Fiction'), (2, 'Science')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	return db
}

func TestBookRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBookRepository(db)
	ctx := context.Background()

	book := &models.Book{
		Title:           "The Go Programming Language",
		Author:          "Alan Donovan",
		ISBN:            "978-0134190440",
		CategoryID:      int64Ptr(1),
		Publisher:       "Addison-Wesley",
		Edition:         "1st",
		PublishedYear:   2015,
		Quantity:        5,
		AvailableCopies: 5,
		Shelf:           "A-12",
		Status:          "available",
		BookURL:         "https://example.com/go-book",
		BookType:        "link",
	}

	err := repo.Create(ctx, book)
	require.NoError(t, err)
	assert.NotZero(t, book.ID)
	assert.False(t, book.CreatedAt.IsZero())

	createdID := book.ID

	got, err := repo.GetByID(ctx, createdID)
	require.NoError(t, err)
	assert.Equal(t, book.Title, got.Title)
	assert.Equal(t, book.Author, got.Author)
	assert.Equal(t, book.ISBN, got.ISBN)
	require.NotNil(t, got.CategoryID)
	assert.Equal(t, int64(1), *got.CategoryID)
	assert.Equal(t, "Fiction", got.CategoryName)
	assert.Equal(t, book.Quantity, got.Quantity)
	assert.Equal(t, book.AvailableCopies, got.AvailableCopies)
	assert.Equal(t, book.Status, got.Status)
	assert.Equal(t, book.BookURL, got.BookURL)
	assert.Equal(t, book.BookType, got.BookType)

	list, err := repo.List(ctx, models.BookFilter{})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, createdID, list[0].ID)

	list, err = repo.List(ctx, models.BookFilter{Search: "Go Programming"})
	require.NoError(t, err)
	assert.Len(t, list, 1)

	list, err = repo.List(ctx, models.BookFilter{Search: "NonExistent"})
	require.NoError(t, err)
	assert.Len(t, list, 0)

	list, err = repo.List(ctx, models.BookFilter{Category: 1})
	require.NoError(t, err)
	assert.Len(t, list, 1)

	list, err = repo.List(ctx, models.BookFilter{Category: 999})
	require.NoError(t, err)
	assert.Len(t, list, 0)

	list, err = repo.List(ctx, models.BookFilter{Author: "Donovan"})
	require.NoError(t, err)
	assert.Len(t, list, 1)

	list, err = repo.List(ctx, models.BookFilter{Status: "available"})
	require.NoError(t, err)
	assert.Len(t, list, 1)

	got.Title = "The Go Programming Language (Updated)"
	got.Quantity = 10
	got.Shelf = "B-05"
	got.Status = "available"
	got.BookURL = "https://example.com/updated"
	got.BookType = "pdf"

	err = repo.Update(ctx, got)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, createdID)
	require.NoError(t, err)
	assert.Equal(t, "The Go Programming Language (Updated)", updated.Title)
	assert.Equal(t, 10, updated.Quantity)
	assert.Equal(t, "B-05", updated.Shelf)
	assert.Equal(t, "https://example.com/updated", updated.BookURL)
	assert.Equal(t, "pdf", updated.BookType)

	err = repo.UpdateAvailability(ctx, createdID, -2)
	require.NoError(t, err)

	afterDelta, err := repo.GetByID(ctx, createdID)
	require.NoError(t, err)
	assert.Equal(t, 3, afterDelta.AvailableCopies)

	err = repo.UpdateAvailability(ctx, createdID, 1)
	require.NoError(t, err)

	afterDelta, err = repo.GetByID(ctx, createdID)
	require.NoError(t, err)
	assert.Equal(t, 4, afterDelta.AvailableCopies)

	err = repo.UpdateBookLink(ctx, createdID, "https://cdn.example.com/ebook.pdf", "epub")
	require.NoError(t, err)

	linked, err := repo.GetByID(ctx, createdID)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/ebook.pdf", linked.BookURL)
	assert.Equal(t, "epub", linked.BookType)

	count, err := repo.CountByCategory(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = repo.CountByCategory(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	err = repo.Delete(ctx, createdID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, createdID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "book not found")

	list, err = repo.List(ctx, models.BookFilter{})
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestBookRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBookRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "book not found")
}

func TestBookRepository_List_MultipleFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBookRepository(db)
	ctx := context.Background()

	b1 := &models.Book{
		Title:           "Clean Code",
		Author:          "Robert Martin",
		ISBN:            "978-0132350884",
		CategoryID:      int64Ptr(1),
		Quantity:        3,
		AvailableCopies: 3,
		Status:          "available",
		BookType:        "link",
	}
	b2 := &models.Book{
		Title:           "The Pragmatic Programmer",
		Author:          "Andrew Hunt",
		ISBN:            "978-0201616224",
		CategoryID:      int64Ptr(2),
		Quantity:        2,
		AvailableCopies: 1,
		Status:          "available",
		BookType:        "pdf",
	}
	require.NoError(t, repo.Create(ctx, b1))
	require.NoError(t, repo.Create(ctx, b2))

	list, err := repo.List(ctx, models.BookFilter{
		Search:   "Code",
		Category: 1,
		Status:   "available",
	})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Clean Code", list[0].Title)
}

func TestBookRepository_Create_Nullables(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBookRepository(db)
	ctx := context.Background()

	book := &models.Book{
		Title:           "Minimal Book",
		Author:          "Unknown",
		CategoryID:      int64Ptr(1),
		Quantity:        1,
		AvailableCopies: 1,
		Status:          "available",
		BookType:        "link",
	}

	err := repo.Create(ctx, book)
	require.NoError(t, err)
	assert.NotZero(t, book.ID)

	got, err := repo.GetByID(ctx, book.ID)
	require.NoError(t, err)
	assert.Equal(t, "", got.ISBN)
	assert.Equal(t, "", got.Publisher)
	assert.Equal(t, "", got.Edition)
	assert.Equal(t, 0, got.PublishedYear)
	assert.Equal(t, "", got.Shelf)
	assert.Equal(t, "", got.BookURL)
	assert.Equal(t, "link", got.BookType)
}