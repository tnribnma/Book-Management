package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"book-management/middleware"
	"book-management/service"
)

type ReservationHandler struct {
	service *service.ReservationService
}

func NewReservationHandler(db *sql.DB) *ReservationHandler {
	return &ReservationHandler{service: service.NewReservationService(db)}
}

func (h *ReservationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		BookID int64 `json:"book_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	reservation, err := h.service.Create(r.Context(), req.BookID, userID)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, reservation)
}