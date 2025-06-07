package editor

import (
	"net/http"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

type signInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type signInResponse struct {
	Token string `json:"token"`
}

// EditorSignInHandler handles editor authentication.
func EditorSignInHandler(svc service.EditorServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req signInRequest
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		signInResp, err := svc.SignIn(r.Context(), req.Email, req.Password)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "editor signin")
			return
		}

		resp := signInResponse{Token: signInResp.IDToken}
		commonHandler.JSONResponse(w, resp, http.StatusOK)
	}
}
