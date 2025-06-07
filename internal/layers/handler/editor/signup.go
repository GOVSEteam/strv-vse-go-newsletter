package editor

import (
	"net/http"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

type signUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
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
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		editor, err := svc.SignUp(r.Context(), req.Email, req.Password)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "editor signup")
			return
		}

		resp := signUpResponse{EditorID: editor.ID, Email: editor.Email}
		commonHandler.JSONResponse(w, resp, http.StatusCreated)
	}
}
