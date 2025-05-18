package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type PasswordResetRequest struct {
	Email string `json:"email"`
}

// POST /editors/password-reset-request
// Triggers Firebase Auth to send a password reset email to the editor.
func FirebasePasswordResetRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req PasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
			return
		}

		apiKey := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

		url := "https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode?key=" + apiKey
		payload := map[string]interface{}{
			"requestType": "PASSWORD_RESET",
			"email":       req.Email,
		}
		jsonPayload, _ := json.Marshal(payload)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			var errResp map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&errResp)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send reset email"})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a reset link has been sent."})
	}
}
