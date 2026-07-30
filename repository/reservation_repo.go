package repository

import (
	"context"
	"database/sql"
	"fmt"
	"book-management/models"
)

type ReservationRepository interface {
	Create(ctx context.Context, reservation *models.Reservation) error
	GetByID(ctx context.Context, id int64) (*models.Reservation, error)
	GetUserReservations(ctx context.Context, userID int64) ([]models.Reservation, error)
	Cancel(ctx context.Context, id int64, userID int64) error
	Fulfill(ctx context.Context, reservationID, bookID int64) error
	HasActiveReservation(ctx context.Context, bookID, userID int64) (bool, error)
}

type reservationRepo struct {
	db *sql.DB
}

func (r *reservationRepo) HasActiveReservation(ctx context.Context, bookID, userID int64) (bool, error) {
	query := `
		SELECT COUNT(*) FROM reservations 
		WHERE book_id = $1 AND user_id = $2 AND status = 'pending'`

	var count int
	err := r.db.QueryRowContext(ctx, query, bookID, userID).Scan(&count)
	return count > 0, err
}

func NewReservationRepository(db *sql.DB) ReservationRepository {
	return &reservationRepo{db: db}
}

func (r *reservationRepo) Create(ctx context.Context, reservation *models.Reservation) error {
	query := `
		INSERT INTO reservations (book_id, user_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, reservation_date`

	err := r.db.QueryRowContext(ctx, query, 
		reservation.BookID, 
		reservation.UserID,
	).Scan(&reservation.ID, &reservation.ReservationDate)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE books 
		SET status = 'reserved' 
		WHERE id = $1`, reservation.BookID)

	if err != nil {
		return fmt.Errorf("failed to update book status: %w", err)
	}

	return nil
}

func (r *reservationRepo) GetByID(ctx context.Context, id int64) (*models.Reservation, error) {
	query := `
		SELECT r.id, r.book_id, bk.title as book_title, r.user_id, u.full_name as user_name,
		       r.reservation_date, r.status, r.created_at
		FROM reservations r
		JOIN books bk ON r.book_id = bk.id
		JOIN users u ON r.user_id = u.id
		WHERE r.id = $1`

	var res models.Reservation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&res.ID, &res.BookID, &res.BookTitle, &res.UserID, &res.UserName,
		&res.ReservationDate, &res.Status, &res.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reservation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return &res, nil
}

func (r *reservationRepo) GetUserReservations(ctx context.Context, userID int64) ([]models.Reservation, error) {
	query := `
		SELECT r.id, r.book_id, bk.title as book_title, r.user_id, 
		       r.reservation_date, r.status, r.created_at
		FROM reservations r
		JOIN books bk ON r.book_id = bk.id
		WHERE r.user_id = $1
		ORDER BY r.reservation_date DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reservations: %w", err)
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var res models.Reservation
		if err := rows.Scan(
			&res.ID, &res.BookID, &res.BookTitle, &res.UserID,
			&res.ReservationDate, &res.Status, &res.CreatedAt,
		); err != nil {
			return nil, err
		}
		reservations = append(reservations, res)
	}

	return reservations, nil
}

func (r *reservationRepo) Cancel(ctx context.Context, id int64, userID int64) error {
	query := `
		UPDATE reservations 
		SET status = 'cancelled'
		WHERE id = $1 AND user_id = $2 AND status = 'pending'`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found or already processed")
	}

	return nil
}

func (r *reservationRepo) Fulfill(ctx context.Context, reservationID, bookID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE reservations 
		SET status = 'fulfilled' 
		WHERE id = $1`, reservationID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE books 
		SET status = 'available' 
		WHERE id = $1`, bookID)
	if err != nil {
		return err
	}

	return tx.Commit()
}