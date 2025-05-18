package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse for consistent error replies
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse for consistent success replies (can be generic or specific)
// For now, we'll often directly marshal the data model (e.g., Newsletter)

// JSONError sends a standard JSON error response.
func JSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: http.StatusText(code), Message: message})
}

// JSONResponse sends a standard JSON success response.
func JSONResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
} 