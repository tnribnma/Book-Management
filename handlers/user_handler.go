package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"book-management/middleware"
	"book-management/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusNotFound, "user not found")
		return
	}
	JSON(w, http.StatusOK, profile)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	JSON(w, http.StatusOK, users)
}

func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role == "" {
		Error(w, http.StatusBadRequest, "role is required")
		return
	}

	user, err := h.service.GetProfile(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.service.UpdateUser(r.Context(), id, user.FullName, req.Role); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}