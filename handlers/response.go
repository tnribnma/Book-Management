package handlers

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Success: false,
		Error:   message,
	}

	json.NewEncoder(w).Encode(response)
}

func ErrorWithDetails(w http.ResponseWriter, status int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Success: false,
		Error:   message,
		Details: details,
	}

	json.NewEncoder(w).Encode(response)
}

func ValidationError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := ErrorResponse{
		Success: false,
		Error:   "validation_error",
		Details: message,
	}

	json.NewEncoder(w).Encode(response)
}