package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"book-management/middleware"
	"book-management/models"
	"book-management/service"
)

type BookHandler struct {
	service *service.BookService
}

func NewBookHandler(db *sql.DB) *BookHandler {
	return &BookHandler{service: service.NewBookService(db)}
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	books, total, err := h.service.List(r.Context(), userID, limit, page)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch books")
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"data": books,
		"meta": map[string]int{"page": page, "limit": limit, "total": total},
	})
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	book, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, book)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	book, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		Error(w, http.StatusNotFound, "book not found")
		return
	}

	JSON(w, http.StatusOK, book)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var req models.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	book, err := h.service.Update(r.Context(), userID, id, req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid book id")
		return
	}

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}