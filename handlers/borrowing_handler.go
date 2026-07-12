package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"book-management/middleware"
	"book-management/models"
	"book-management/service"
	"book-management/utils"
)

type BorrowingHandler struct {
	service *service.BorrowingService
}

func NewBorrowingHandler(db *sql.DB) *BorrowingHandler {
	return &BorrowingHandler{service: service.NewBorrowingService(db)}
}

func (h *BorrowingHandler) IssueBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.BorrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request")
		return
	}

	borrowing, err := h.service.Borrow(r.Context(), userID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, borrowing)
}

func (h *BorrowingHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	var req models.ReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := h.service.Return(r.Context(), req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Book returned successfully"})
}

func (h *BorrowingHandler) GetMyBorrowings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	borrowings, err := h.service.GetUserBorrowings(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, borrowings)
}