package repository

import (
	"context"
	"database/sql"
	"book-management/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user models.User, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		user.Email, passwordHash, user.FullName, "user").Scan(&id)
	return id, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, string, error) {
	var user models.User
	var passwordHash string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, full_name, role, created_at, password_hash 
		FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Role, &user.CreatedAt, &passwordHash)
	return user, passwordHash, err
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, full_name, role, created_at 
		FROM users WHERE id = $1`, id).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Role, &user.CreatedAt)
	return user, err
}

func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, full_name, role, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, id)
	return err
}