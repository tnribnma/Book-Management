package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"book-management/config"
	"book-management/utils"
)

type AuthHandler struct {
	DB  *sql.DB
	JWT *config.JWTConfig
}

func NewAuthHandler(db *sql.DB, jwt *config.JWTConfig) *AuthHandler {
	return &AuthHandler{DB: db, JWT: jwt}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var id int64
	err = h.DB.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		req.Email, hash,
	).Scan(&id)

	if err != nil {
		Error(w, http.StatusBadRequest, "email already exists")
		return
	}

	JSON(w, http.StatusCreated, map[string]any{"id": id, "email": req.Email})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var user struct {
		ID           int64
		PasswordHash string
	}

	err := h.DB.QueryRow(
		`SELECT id, password_hash FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.PasswordHash)

	if err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := utils.CheckPassword(user.PasswordHash, req.Password); err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := utils.CreateToken(user.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	JSON(w, http.StatusOK, map[string]any{"token": token})
}