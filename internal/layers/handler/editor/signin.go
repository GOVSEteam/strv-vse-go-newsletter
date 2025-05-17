package editor

import (
	"encoding/json"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"net/http"
)

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func EditorSignInHandler(svc service.EditorService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req signInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp, err := svc.SignIn(req.Email, req.Password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
