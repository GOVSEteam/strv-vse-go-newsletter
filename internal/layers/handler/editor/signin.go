package editor

import (
	"encoding/json"
	"fmt"
	"net/http"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// EditorSignInHandler handles editor login.
func EditorSignInHandler(svc service.EditorServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req signInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Basic validation for required fields. More complex validation is handled by the EditorService.
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

		resp, err := svc.SignIn(r.Context(), req.Email, req.Password)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// log.Printf("Error editor sign in: %v", err)
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		commonHandler.JSONResponse(w, resp, http.StatusOK)
	}
}
