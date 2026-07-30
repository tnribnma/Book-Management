package handlers

import (
	"encoding/json"
	"net/http"

	"book-management/middleware"
	"book-management/service"
)

type BorrowingHandler struct {
	service service.BorrowingService
}

func NewBorrowingHandler(service service.BorrowingService) *BorrowingHandler {
	return &BorrowingHandler{	service: service}
}

func (h *BorrowingHandler) IssueBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		BookID int64 `json:"book_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := h.service.IssueBook(r.Context(), req.BookID, userID); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, map[string]string{"message": "Book issued successfully"})
}

func (h *BorrowingHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		BookID int64 `json:"book_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := h.service.ReturnBook(r.Context(), req.BookID, userID); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "Book returned successfully"})
}

func (h *BorrowingHandler) GetMyBorrowings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		Error(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	borrowings, err := h.service.GetMyBorrowings(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, borrowings)
}