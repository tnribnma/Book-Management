package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"book-management/models"
	"book-management/service"
	"book-management/utils"
	"book-management/validators" 
)

type BookHandler struct {
	service *service.BookService
}

func NewBookHandler(db *sql.DB) *BookHandler {
	return &BookHandler{
		service: service.NewBookService(db),
	}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req models.BookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := validators.Validate.Struct(req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	book, err := h.service.Create(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, book)  
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	filter := models.BookFilter{
		Search:   r.URL.Query().Get("search"),
		Author:   r.URL.Query().Get("author"),
		Status:   r.URL.Query().Get("status"),
	}

	categoryStr := r.URL.Query().Get("category_id")
	if categoryStr != "" {
		if catID, err := strconv.ParseInt(categoryStr, 10, 64); err == nil {
			filter.Category = catID
		}
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	books, total, err := h.service.List(r.Context(), filter, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch books")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"data": books,
		"meta": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	book, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Book not found")
		return
	}

	utils.JSON(w, http.StatusOK, book)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	var req models.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	book, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
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

	books, err := h.service.Search(r.Context(), query)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Search failed")
		return
	}

	utils.JSON(w, http.StatusOK, books)
}