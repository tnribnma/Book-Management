
package repository

import (
	"book-management/models"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUserRepoMock(t *testing.T) (UserRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewUserRepository(db)
	cleanup := func() { db.Close() }
	return repo, mock, cleanup
}

func TestUserRepo_Create_Success(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	user := &models.User{
		Email:        "alice@example.com",
		PasswordHash: "hashed",
		FullName:     "Alice",
		Role:         "user",
	}
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash, user.FullName, user.Role).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

	err := repo.Create(ctx, user)

	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, now, user.CreatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Create_Error(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	user := &models.User{
		Email:        "alice@example.com",
		PasswordHash: "hashed",
		FullName:     "Alice",
		Role:         "user",
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash, user.FullName, user.Role).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByEmail_Success(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "full_name", "role", "created_at"}).
		AddRow(1, "alice@example.com", "hashed", "Alice", "user", now)

	mock.ExpectQuery(`SELECT id, email, password_hash, full_name, role, created_at\s+FROM users\s+WHERE email = \$1`).
		WithArgs("alice@example.com").
		WillReturnRows(rows)

	user, err := repo.GetByEmail(ctx, "alice@example.com")

	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "hashed", user.PasswordHash)
	assert.Equal(t, "Alice", user.FullName)
	assert.Equal(t, "user", user.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, email, password_hash, full_name, role, created_at\s+FROM users\s+WHERE email = \$1`).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByEmail(ctx, "missing@example.com")

	assert.Nil(t, user)
	assert.EqualError(t, err, "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByEmail_DBError(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, email, password_hash, full_name, role, created_at\s+FROM users\s+WHERE email = \$1`).
		WithArgs("alice@example.com").
		WillReturnError(errors.New("connection refused"))

	user, err := repo.GetByEmail(ctx, "alice@example.com")

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "failed to get user by email")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "full_name", "role", "created_at"}).
		AddRow(1, "alice@example.com", "hashed", "Alice", "user", now)

	mock.ExpectQuery(`SELECT id, email, password_hash, full_name, role, created_at\s+FROM users\s+WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	user, err := repo.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, email, password_hash, full_name, role, created_at\s+FROM users\s+WHERE id = \$1`).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID(ctx, 99)

	assert.Nil(t, user)
	assert.EqualError(t, err, "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByID_DBError(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, email, password_hash, full_name, role, created_at\s+FROM users\s+WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("timeout"))

	user, err := repo.GetByID(ctx, 1)

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "failed to get user by ID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_List_Success(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "full_name", "role", "created_at"}).
		AddRow(2, "bob@example.com", "Bob", "librarian", now).
		AddRow(1, "alice@example.com", "Alice", "user", now.Add(-time.Hour))

	mock.ExpectQuery(`SELECT id, email, full_name, role, created_at\s+FROM users\s+ORDER BY created_at DESC`).
		WillReturnRows(rows)

	users, err := repo.List(ctx)

	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "bob@example.com", users[0].Email)
	assert.Equal(t, "alice@example.com", users[1].Email)
	assert.Empty(t, users[0].PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_List_Empty(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "email", "full_name", "role", "created_at"})
	mock.ExpectQuery(`SELECT id, email, full_name, role, created_at\s+FROM users\s+ORDER BY created_at DESC`).
		WillReturnRows(rows)

	users, err := repo.List(ctx)

	require.NoError(t, err)
	assert.Empty(t, users)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_List_QueryError(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, email, full_name, role, created_at\s+FROM users\s+ORDER BY created_at DESC`).
		WillReturnError(errors.New("db down"))

	users, err := repo.List(ctx)

	assert.Nil(t, users)
	assert.ErrorContains(t, err, "failed to list users")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_List_ScanError(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "email", "full_name", "role", "created_at"}).
		AddRow("not-an-int", "alice@example.com", "Alice", "user", time.Now())

	mock.ExpectQuery(`SELECT id, email, full_name, role, created_at\s+FROM users\s+ORDER BY created_at DESC`).
		WillReturnRows(rows)

	users, err := repo.List(ctx)

	assert.Nil(t, users)
	assert.ErrorContains(t, err, "failed to scan user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Update_Success(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	user := &models.User{ID: 1, FullName: "Alice Updated", Role: "librarian"}

	mock.ExpectExec(`UPDATE users\s+SET full_name = \$1, role = \$2\s+WHERE id = \$3`).
		WithArgs(user.FullName, user.Role, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Update_NotFound(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	user := &models.User{ID: 99, FullName: "Nobody", Role: "user"}

	mock.ExpectExec(`UPDATE users\s+SET full_name = \$1, role = \$2\s+WHERE id = \$3`).
		WithArgs(user.FullName, user.Role, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(ctx, user)

	assert.EqualError(t, err, "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Update_DBError(t *testing.T) {
	repo, mock, cleanup := newUserRepoMock(t)
	defer cleanup()
	ctx := context.Background()

	user := &models.User{ID: 1, FullName: "Alice", Role: "admin"}

	mock.ExpectExec(`UPDATE users\s+SET full_name = \$1, role = \$2\s+WHERE id = \$3`).
		WithArgs(user.FullName, user.Role, user.ID).
		WillReturnError(errors.New("constraint violation"))

	err := repo.Update(ctx, user)

	assert.ErrorContains(t, err, "failed to update user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_UpdateRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepo{db: db}
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET role = \$1 WHERE id = \$2`).
		WithArgs("admin", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateRole(ctx, 1, "admin")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_UpdateRole_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &userRepo{db: db}
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET role = \$1 WHERE id = \$2`).
		WithArgs("admin", int64(1)).
		WillReturnError(errors.New("db error"))

	err = repo.UpdateRole(ctx, 1, "admin")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
