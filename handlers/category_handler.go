package handlers

import (
	"encoding/json"
	"net/http"
	"book-management/service"
)

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{ service: service}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	category, err := h.service.CreateCategory(r.Context(), req.Name, req.Description)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	JSON(w, http.StatusOK, categories)
}