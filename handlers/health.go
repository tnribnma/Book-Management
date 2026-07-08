package handlers

import "net/http"

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
	})
}