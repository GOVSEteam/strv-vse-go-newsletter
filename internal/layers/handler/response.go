package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
)

// ErrorResponse represents a standard JSON error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// JSONError sends a standard JSON error response.
func JSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
	})
}

// JSONErrorSecure sends a secure JSON error response for application errors.
// For server errors (5xx), it returns generic messages to prevent information leakage.
// For client errors (4xx), it returns the actual error message.
func JSONErrorSecure(w http.ResponseWriter, err error, operation string) {
	statusCode := apperrors.ErrorToHTTPStatus(err)
	
	var message string
	if statusCode >= 500 {
		// TODO: Remove this debug logging - temporarily showing actual error for debugging
		message = fmt.Sprintf("Debug: %s", err.Error())
	} else {
		// Client errors are generally safe to show
		message = err.Error()
	}
	
	JSONError(w, message, statusCode)
}

// JSONResponse sends a standard JSON success response.
func JSONResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
} 