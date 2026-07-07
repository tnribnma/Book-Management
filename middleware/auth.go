package middleware

import (
	"context"
	"net/http"
	"strings"

	"book-management/utils"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

func Auth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		userID, err := utils.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) int64 {
	if v := r.Context().Value(UserIDKey); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}