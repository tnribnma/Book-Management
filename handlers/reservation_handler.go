package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"book-management/middleware"
	"book-management/service"
)

type ReservationHandler struct {
	service service.ReservationService
}

func NewReservationHandler(service service.ReservationService) *ReservationHandler {
	return &ReservationHandler{	service: service}
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	reservation, err := h.service.CreateReservation(r.Context(), req.BookID, userID)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, reservation)
}

func (h *ReservationHandler) GetMyReservations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	reservations, err := h.service.GetUserReservations(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, reservations)
}

func (h *ReservationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid reservation ID")
		return
	}

	if err := h.service.CancelReservation(r.Context(), id, userID); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "reservation cancelled"})
}