package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"book-management/middleware"
	"book-management/models"
	"book-management/service"
	"book-management/validators"
)

type BookHandler struct {
	service service.BookService
}

func NewBookHandler(service service.BookService) *BookHandler {
	return &BookHandler{service: service}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req models.BookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := validators.Validate.Struct(req); err != nil {
		Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	book, err := h.service.CreateBook(r.Context(), req, userID)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, book)
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	filter := models.BookFilter{
		Search: r.URL.Query().Get("search"),
		Author: r.URL.Query().Get("author"),
		Status: r.URL.Query().Get("status"),
	}

	categoryStr := r.URL.Query().Get("category_id")
	if categoryStr != "" {
		if catID, err := strconv.ParseInt(categoryStr, 10, 64); err == nil {
			filter.Category = catID
		}
	}

	books, err := h.service.ListBooks(r.Context(), filter)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch books")
		return
	}

	JSON(w, http.StatusOK, books)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	book, err := h.service.GetBook(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "Book not found")
		return
	}

	JSON(w, http.StatusOK, book)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	var req models.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	book, err := h.service.UpdateBook(r.Context(), id, req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	if err := h.service.DeleteBook(r.Context(), id); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.ListBooks(w, r)
		return
	}

	filter := models.BookFilter{Search: query}
	books, err := h.service.ListBooks(r.Context(), filter)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Search failed")
		return
	}

	JSON(w, http.StatusOK, books)
}