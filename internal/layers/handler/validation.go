package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationResponse represents a simple validation error response.
type ValidationResponse struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// ValidateAndRespond decodes JSON and validates it, sending error response if needed.
// Returns true if validation passed, false if validation failed (response already sent).
func ValidateAndRespond(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		sendValidationError(w, []string{"Invalid JSON format"})
		return false
	}

	// Validate struct
	validate := validator.New()
	if err := validate.Struct(v); err != nil {
		if validatorErrs, ok := err.(validator.ValidationErrors); ok {
			var errors []string
			for _, fieldErr := range validatorErrs {
				errors = append(errors, getValidationMessage(fieldErr))
			}
			sendValidationError(w, errors)
		} else {
			sendValidationError(w, []string{"Validation failed"})
		}
		return false
	}

	return true
}

// sendValidationError sends a validation error response.
func sendValidationError(w http.ResponseWriter, errors []string) {
	response := ValidationResponse{
		Message: "Validation failed",
		Errors:  errors,
	}
	
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

// getValidationMessage returns user-friendly messages for newsletter service validation rules.
func getValidationMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be no more than %s characters long", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
} 