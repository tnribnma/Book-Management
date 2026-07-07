package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth_MissingToken(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d",
			http.StatusUnauthorized,
			w.Code)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "abc123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d",
			http.StatusUnauthorized,
			w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d",
			http.StatusUnauthorized,
			w.Code)
	}
}

func TestGetUserID_NoUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id := GetUserID(req)
	if id != 0 {
		t.Errorf("expected 0, got %d", id)
	}
}

func TestGetUserID_WithUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(
		req.Context(),
		UserIDKey,
		int64(123),
	)
	req = req.WithContext(ctx)
	id := GetUserID(req)
	if id != 123 {
		t.Errorf("expected 123, got %d", id)
	}
}
