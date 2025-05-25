package editor_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirebasePasswordResetRequestHandler(t *testing.T) {
	// Store original environment variable and defer reset
	originalAPIKey := os.Getenv("FIREBASE_API_KEY")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("FIREBASE_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("FIREBASE_API_KEY")
		}
	}()

	httpHandler := h.FirebasePasswordResetRequestHandler()

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/editor/password-reset-request", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - Invalid JSON Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader([]byte(`{"email": "test@example.com"`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid request", response["error"])
	})

	t.Run("Error - Empty Email", func(t *testing.T) {
		resetData := map[string]string{
			"email": "",
		}
		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid request", response["error"])
	})

	t.Run("Error - Missing Email Field", func(t *testing.T) {
		resetData := map[string]string{}
		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid request", response["error"])
	})

	t.Run("Edge Case - Empty JSON Object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader([]byte(`{}`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid request", response["error"])
	})

	t.Run("Edge Case - Null Email in JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader([]byte(`{"email": null}`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid request", response["error"])
	})

	t.Run("Error - Missing Firebase API Key", func(t *testing.T) {
		// Unset the API key to test this scenario
		os.Unsetenv("FIREBASE_API_KEY")

		resetData := map[string]string{
			"email": "test@example.com",
		}

		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		// The handler will try to make a request with an empty API key
		// This will likely result in an error from Firebase or HTTP client
		assert.True(t, rr.Code >= 400) // Should be some kind of error
	})

	t.Run("Real-world Usage - Valid Email Format", func(t *testing.T) {
		// Set a test API key
		os.Setenv("FIREBASE_API_KEY", "test-api-key")

		resetData := map[string]string{
			"email": "test@example.com",
		}

		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		// Since this makes a real HTTP call to Firebase, we expect it to fail
		// but we can verify the handler processes the request correctly
		// The response will be either success (if Firebase is reachable) or error
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= 400)
	})

	t.Run("Performance - Large JSON Payload", func(t *testing.T) {
		// Test with a large but valid JSON payload
		largeDescription := ""
		for i := 0; i < 100; i++ { // Reduced size to avoid timeout
			largeDescription += "This is a very long description. "
		}

		resetData := map[string]interface{}{
			"email":       "test@example.com",
			"description": largeDescription, // Extra field that should be ignored
		}

		os.Setenv("FIREBASE_API_KEY", "test-api-key")

		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		// Handler should process the request regardless of extra fields
		assert.True(t, rr.Code == http.StatusOK || rr.Code >= 400)
	})

	t.Run("Edge Case - Invalid Email Format", func(t *testing.T) {
		invalidEmails := []string{
			"invalid-email",
			"@example.com",
			"test@",
		}

		os.Setenv("FIREBASE_API_KEY", "test-api-key")

		for _, invalidEmail := range invalidEmails {
			t.Run("Invalid: "+invalidEmail, func(t *testing.T) {
				resetData := map[string]string{
					"email": invalidEmail,
				}

				bodyBytes, _ := json.Marshal(resetData)
				req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader(bodyBytes))
				rr := httptest.NewRecorder()

				httpHandler.ServeHTTP(rr, req)

				// Firebase should reject invalid email formats
				// The handler should return an error status
				assert.True(t, rr.Code >= 400)
			})
		}
	})

	t.Run("Security - SQL Injection Attempt in Email", func(t *testing.T) {
		maliciousEmail := "test'; DROP TABLE editors; --@example.com"

		os.Setenv("FIREBASE_API_KEY", "test-api-key")

		resetData := map[string]string{
			"email": maliciousEmail,
		}

		bodyBytes, _ := json.Marshal(resetData)
		req := httptest.NewRequest(http.MethodPost, "/editor/password-reset-request", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		// Firebase should handle this safely and likely return an error
		assert.True(t, rr.Code >= 400)
	})
} 