package handlers

import (
	"log"
	"encoding/json"
	"net/http"
	"book-management/models"
	"book-management/utils"
	"book-management/service"
)

type AuthHandler struct {
	userService service.UserService
}

func NewAuthHandler(userService service.UserService) *AuthHandler { 
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Register called")

	var req models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Decode error:", err)
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	log.Printf("Request: %+v\n", req)

	user, err := h.userService.Register(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		log.Println("Register error:", err)
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Println("User created:", user.Email)

	JSON(w, http.StatusCreated, map[string]any{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := utils.CreateToken(user.ID, user.Role)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}