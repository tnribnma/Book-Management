package middleware

import (
	"net/http"

	"book-management/utils"
)

func Role(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r)

			for _, allowed := range allowedRoles {
				if userRole == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			if userRole == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			utils.Error(w, http.StatusForbidden, "Insufficient permissions")
		}
	}
}