package editor

import (
	"errors"
	"net/http"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetRequestHandler handles requests to initiate a password reset for an editor.
// It uses the PasswordResetService to send the reset email.
func PasswordResetRequestHandler(pwdResetSvc service.PasswordResetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PasswordResetRequest
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		if err := pwdResetSvc.SendPasswordResetEmail(r.Context(), req.Email); err != nil {
			// Important: Do not reveal if the email exists or not for non-client errors.
			// If the error is a validation error (e.g. invalid email format from service), show it.
			if errors.Is(err, apperrors.ErrValidation) || errors.Is(err, apperrors.ErrBadRequest) {
				commonHandler.JSONErrorSecure(w, err, "password reset validation")
			} else {
				// For other errors (including not found, internal), still return a generic success-like response
				// to prevent leaking information about email existence. Log the actual error server-side.
				commonHandler.JSONErrorSecure(w, err, "password reset")
				commonHandler.JSONResponse(w, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."}, http.StatusOK)
			}
			return
		}

		// Always return a generic message to prevent email enumeration.
		commonHandler.JSONResponse(w, map[string]string{"message": "If an account with that email exists, a password reset link has been sent."}, http.StatusOK)
	}
}
