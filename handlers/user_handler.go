package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"book-management/middleware"
	"book-management/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{service: service.NewUserService(nil)} 
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
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var req struct {
		Role string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	err := h.service.UpdateRole(r.Context(), id, req.Role)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}