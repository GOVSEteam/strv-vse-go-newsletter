package editor

import (
	"encoding/json"
	"fmt"
	"net/http"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpResponse struct {
	EditorID string `json:"editor_id"`
	Email    string `json:"email"`
	// Could also include CreatedAt from models.Editor if desired
}

// EditorSignUpHandler handles new editor registration.
func EditorSignUpHandler(svc service.EditorServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req signUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Basic validation for required fields. More complex validation (e.g. email format, pass length)
		// is handled by the EditorService.
		if req.Email == "" {
			err := fmt.Errorf("email cannot be empty: %w", apperrors.ErrValidation)
			commonHandler.JSONError(w, err.Error(), apperrors.ErrorToHTTPStatus(err))
			return
		}
		if req.Password == "" {
			err := fmt.Errorf("password cannot be empty: %w", apperrors.ErrValidation)
			commonHandler.JSONError(w, err.Error(), apperrors.ErrorToHTTPStatus(err))
			return
		}

		editor, err := svc.SignUp(r.Context(), req.Email, req.Password)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// log.Printf("Error editor sign up: %v", err)
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		resp := signUpResponse{EditorID: editor.ID, Email: editor.Email}
		commonHandler.JSONResponse(w, resp, http.StatusCreated)
	}
}
