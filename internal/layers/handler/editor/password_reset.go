package editor

import (
	// "bytes" // No longer needed
	"encoding/json"
	"fmt"
	"net/http"

	// "os" // No longer needed

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetRequestHandler handles requests to initiate a password reset for an editor.
// It uses the PasswordResetService to send the reset email.
func PasswordResetRequestHandler(pwdResetSvc service.PasswordResetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.Email == "" {
			err := fmt.Errorf("email cannot be empty: %w", apperrors.ErrValidation)
			commonHandler.JSONError(w, err.Error(), apperrors.ErrorToHTTPStatus(err))
			return
		}

		if err := pwdResetSvc.SendPasswordResetEmail(r.Context(), req.Email); err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// Important: Do not reveal if the email exists or not for non-client errors.
			// The service might return ErrEditorNotFound, but we typically map this to a generic success
			// or a less specific client error for password reset flows to prevent account enumeration.
			// However, if ErrorToHTTPStatus maps ErrNotFound to 404, and we want to avoid that here,
			// we might need custom logic. For now, let's assume service returns appropriate errors
			// that ErrorToHTTPStatus will handle correctly (e.g. validation errors as 400, internal as 500).
			// If the service returns ErrEditorNotFound, and ErrorToHTTPStatus maps it to 404, this leaks info.
			// A common pattern is to always return OK for password resets unless it's a clear client malformed request.

			// If the error is a validation error (e.g. invalid email format from service), show it.
			if apperrors.IsValidation(err) || apperrors.IsBadRequest(err) {
				commonHandler.JSONError(w, err.Error(), statusCode)
			} else {
				// For other errors (including not found, internal), still return a generic success-like response
				// to prevent leaking information about email existence. Log the actual error server-side.
				// log.Printf("Password reset internal processing error for email %s: %v", req.Email, err)
				commonHandler.JSONResponse(w, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."}, http.StatusOK)
			}
			return
		}

		// Always return a generic message to prevent email enumeration.
		commonHandler.JSONResponse(w, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."}, http.StatusOK)
	}
}
