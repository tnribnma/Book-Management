package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"book-management/models"
	"book-management/service"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(db *sql.DB) *CategoryHandler {
	return &CategoryHandler{service: service.NewCategoryService(db)}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.Category
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	category, err := h.service.Create(r.Context(), req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	JSON(w, http.StatusOK, categories)
}