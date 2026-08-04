package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"book-management/models"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBorrowingTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := "postgres://postgres:1234@localhost:5432/books_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	_, err = db.Exec(`
		TRUNCATE borrowings, books, categories, users RESTART IDENTITY CASCADE;
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO categories (id, name) VALUES (1, 'Fiction')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO users (id, full_name, email, password_hash, role)
		VALUES (1, 'Test User', 'test@example.com', 'hash', 'user')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO books (id, title, author, category_id, quantity, available_copies, status, book_type)
		VALUES (1, 'Test Book', 'Test Author', 1, 5, 5, 'available', 'link')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	return db
}

func TestBorrowingRepository_IssueAndReturn(t *testing.T) {
	db := setupBorrowingTestDB(t)
	defer db.Close()

	repo := NewBorrowingRepository(db)
	ctx := context.Background()

	dueDate := time.Now().Add(7 * 24 * time.Hour) 

	borrowing := &models.Borrowing{
		BookID:  1,
		UserID:  1,
		DueDate: dueDate,
	}

	err := repo.IssueBook(ctx, borrowing)
	require.NoError(t, err)
	assert.NotZero(t, borrowing.ID)
	assert.False(t, borrowing.IssueDate.IsZero())

	var available int
	err = db.QueryRow(`SELECT available_copies FROM books WHERE id = 1`).Scan(&available)
	require.NoError(t, err)
	assert.Equal(t, 4, available)

	has, err := repo.HasActiveBorrowing(ctx, 1, 1)
	require.NoError(t, err)
	assert.True(t, has)

	has, err = repo.HasActiveBorrowing(ctx, 1, 999)
	require.NoError(t, err)
	assert.False(t, has)

	count, err := repo.CountActiveBorrowings(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	list, err := repo.GetMyBorrowings(ctx, 1)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), list[0].BookID)
	assert.Equal(t, "Test Book", list[0].BookTitle)
	assert.Equal(t, "borrowed", list[0].Status)

	err = repo.ReturnBook(ctx, 1, 1)
	require.NoError(t, err)

	err = db.QueryRow(`SELECT available_copies FROM books WHERE id = 1`).Scan(&available)
	require.NoError(t, err)
	assert.Equal(t, 5, available)

	has, err = repo.HasActiveBorrowing(ctx, 1, 1)
	require.NoError(t, err)
	assert.False(t, has)

	count, err = repo.CountActiveBorrowings(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestBorrowingRepository_ReturnBook_NotFound(t *testing.T) {
	db := setupBorrowingTestDB(t)
	defer db.Close()

	repo := NewBorrowingRepository(db)
	ctx := context.Background()

	err := repo.ReturnBook(ctx, 1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active borrowing found")
}

func TestBorrowingRepository_Overdue(t *testing.T) {
	db := setupBorrowingTestDB(t)
	defer db.Close()

	repo := NewBorrowingRepository(db)
	ctx := context.Background()

	pastDue := time.Now().Add(-3 * 24 * time.Hour) 

	borrowing := &models.Borrowing{
		BookID:  1,
		UserID:  1,
		DueDate: pastDue,
	}
	err := repo.IssueBook(ctx, borrowing)
	require.NoError(t, err)

	has, err := repo.HasOverdueBorrowing(ctx, 1)
	require.NoError(t, err)
	assert.True(t, has)

	has, err = repo.HasOverdueBorrowing(ctx, 999)
	require.NoError(t, err)
	assert.False(t, has)

	overdueList, err := repo.GetOverdueBorrowings(ctx)
	require.NoError(t, err)
	require.Len(t, overdueList, 1)
	assert.Equal(t, int64(1), overdueList[0].BookID)
	assert.Equal(t, "Test Book", overdueList[0].BookTitle)
	assert.Equal(t, "Test User", overdueList[0].UserName)
	assert.Equal(t, "borrowed", overdueList[0].Status)

	err = repo.ReturnBook(ctx, 1, 1)
	require.NoError(t, err)

	overdueList, err = repo.GetOverdueBorrowings(ctx)
	require.NoError(t, err)
	assert.Len(t, overdueList, 0)

	has, err = repo.HasOverdueBorrowing(ctx, 1)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestBorrowingRepository_MultipleBorrowings(t *testing.T) {
	db := setupBorrowingTestDB(t)
	defer db.Close()

	repo := NewBorrowingRepository(db)
	ctx := context.Background()

	_, err := db.Exec(`
		INSERT INTO books (id, title, author, category_id, quantity, available_copies, status, book_type)
		VALUES (2, 'Second Book', 'Author 2', 1, 3, 3, 'available', 'pdf')
	`)
	require.NoError(t, err)

	dueDate := time.Now().Add(14 * 24 * time.Hour)

	b1 := &models.Borrowing{BookID: 1, UserID: 1, DueDate: dueDate}
	b2 := &models.Borrowing{BookID: 2, UserID: 1, DueDate: dueDate}

	require.NoError(t, repo.IssueBook(ctx, b1))
	require.NoError(t, repo.IssueBook(ctx, b2))

	count, err := repo.CountActiveBorrowings(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	list, err := repo.GetMyBorrowings(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, list, 2)

	require.NoError(t, repo.ReturnBook(ctx, 1, 1))

	count, err = repo.CountActiveBorrowings(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}