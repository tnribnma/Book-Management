package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"book-management/utils"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")

		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		userID, err := utils.ParseToken(tokenStr, os.Getenv("JWT_SECRET"))
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) int64 {
	v := r.Context().Value(UserIDKey)
	if v == nil {
		return 0
	}
	return v.(int64)
}
