package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"book-management/middleware"
)

type BookHandler struct {
	DB *sql.DB
}

func NewBookHandler(db *sql.DB) *BookHandler {
	return &BookHandler{DB: db}
}

type bookReq struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedYear int    `json:"published_year"`
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	rows, err := h.DB.Query(`SELECT id, user_id, title, author, published_year, created_at FROM books WHERE user_id=$1 ORDER BY id DESC`, userID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type book struct {
		ID            int64  `json:"id"`
		UserID        int64  `json:"user_id"`
		Title         string `json:"title"`
		Author        string `json:"author"`
		PublishedYear int    `json:"published_year"`
		CreatedAt     string `json:"created_at"`
	}

	var books []book
	for rows.Next() {
		var b book
		if err := rows.Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.PublishedYear, &b.CreatedAt); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req bookReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var id int64
	err := h.DB.QueryRow(
		`INSERT INTO books (user_id, title, author, published_year) VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, req.Title, req.Author, req.PublishedYear,
	).Scan(&id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	idStr := strings.TrimPrefix(r.URL.Path, "/books/")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var b bookReq
	var uid int64
	err := h.DB.QueryRow(
		`SELECT user_id, title, author, published_year FROM books WHERE id=$1`,
		id,
	).Scan(&uid, &b.Title, &b.Author, &b.PublishedYear)
	if err != nil || uid != userID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"id": id, "title": b.Title, "author": b.Author, "published_year": b.PublishedYear})
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "implement update", http.StatusNotImplemented)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "implement delete", http.StatusNotImplemented)
}
