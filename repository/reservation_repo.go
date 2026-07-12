package repository

import (
	"context"
	"database/sql"
	"book-management/models"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(ctx context.Context, reservation models.Reservation) (models.Reservation, error) {
	var created models.Reservation
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO reservations (user_id, book_id, status, reservation_date)
		VALUES ($1, $2, 'pending', NOW())
		RETURNING id, user_id, book_id, status, reservation_date`,
		reservation.UserID, reservation.BookID).Scan(
		&created.ID,
		&created.UserID,
		&created.BookID,
		&created.Status,
		&created.ReservationDate,
	)
	return created, err
}

func (r *ReservationRepository) List(ctx context.Context, userID int64) ([]models.Reservation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, book_id, status, reservation_date 
		FROM reservations 
		WHERE user_id = $1 
		ORDER BY reservation_date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var res models.Reservation
		err := rows.Scan(
			&res.ID,
			&res.UserID,
			&res.BookID,
			&res.Status,
			&res.ReservationDate,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}