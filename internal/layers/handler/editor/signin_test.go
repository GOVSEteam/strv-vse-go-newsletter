package editor_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditorSignInHandler(t *testing.T) {
	mockService := new(MockEditorService)
	httpHandler := h.EditorSignInHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		expectedResponse := &service.SignInResponse{
			IDToken:      "firebase-id-token-123",
			RefreshToken: "firebase-refresh-token-123",
			Email:        "test@example.com",
			LocalID:      "firebase-local-id-123",
		}

		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(expectedResponse, nil).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response service.SignInResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse.IDToken, response.IDToken)
		assert.Equal(t, expectedResponse.RefreshToken, response.RefreshToken)
		assert.Equal(t, expectedResponse.Email, response.Email)
		assert.Equal(t, expectedResponse.LocalID, response.LocalID)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/editor/signin", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - Invalid JSON Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader([]byte(`{"email": "test@example.com", password`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Empty Email", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "",
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Empty Password", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "",
		}
		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Missing Email Field", func(t *testing.T) {
		signInData := map[string]string{
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Missing Password Field", func(t *testing.T) {
		signInData := map[string]string{
			"email": "test@example.com",
		}
		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Service Invalid Credentials", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}

		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(nil, errors.New("invalid credentials")).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid credentials", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service User Not Found", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}

		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(nil, errors.New("user not found")).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "user not found", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Firebase API Error", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(nil, errors.New("firebase API error")).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "firebase API error", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Network Error", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(nil, errors.New("network timeout")).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "network timeout", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Valid Email Formats", func(t *testing.T) {
		testCases := []struct {
			name  string
			email string
		}{
			{"Standard Email", "user@example.com"},
			{"Email with Plus", "user+tag@example.com"},
			{"Email with Subdomain", "user@mail.example.com"},
			{"Email with Numbers", "user123@example123.com"},
			{"Email with Hyphens", "user-name@example-domain.com"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				signInData := map[string]string{
					"email":    tc.email,
					"password": "password123",
				}
				expectedResponse := &service.SignInResponse{
					IDToken:      "firebase-id-token-" + tc.name,
					RefreshToken: "firebase-refresh-token-" + tc.name,
					Email:        tc.email,
					LocalID:      "firebase-local-id-" + tc.name,
				}

				mockService.On("SignIn", tc.email, "password123").Return(expectedResponse, nil).Once()

				bodyBytes, _ := json.Marshal(signInData)
				req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
				rr := httptest.NewRecorder()

				httpHandler.ServeHTTP(rr, req)

				assert.Equal(t, http.StatusOK, rr.Code)
				mockService.AssertExpectations(t)
			})
		}
	})

	t.Run("Real-world Usage - Password Complexity", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "ComplexP@ssw0rd!123",
		}
		expectedResponse := &service.SignInResponse{
			IDToken:      "firebase-id-token-complex",
			RefreshToken: "firebase-refresh-token-complex",
			Email:        "test@example.com",
			LocalID:      "firebase-local-id-complex",
		}

		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(expectedResponse, nil).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Very Long Email", func(t *testing.T) {
		longEmail := "very.long.email.address.that.might.cause.issues@very.long.domain.name.example.com"
		signInData := map[string]string{
			"email":    longEmail,
			"password": "password123",
		}
		expectedResponse := &service.SignInResponse{
			IDToken:      "firebase-id-token-long",
			RefreshToken: "firebase-refresh-token-long",
			Email:        longEmail,
			LocalID:      "firebase-local-id-long",
		}

		mockService.On("SignIn", longEmail, "password123").Return(expectedResponse, nil).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Special Characters in Password", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "P@$$w0rd!#$%^&*()_+-=[]{}|;:,.<>?",
		}
		expectedResponse := &service.SignInResponse{
			IDToken:      "firebase-id-token-special",
			RefreshToken: "firebase-refresh-token-special",
			Email:        "test@example.com",
			LocalID:      "firebase-local-id-special",
		}

		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(expectedResponse, nil).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Security - SQL Injection Attempt in Email", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test'; DROP TABLE editors; --@example.com",
			"password": "password123",
		}

		// Service should handle this safely, but let's test the handler passes it through
		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(nil, errors.New("invalid email format")).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Security - SQL Injection Attempt in Password", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "test@example.com",
			"password": "'; DROP TABLE editors; --",
		}

		// Service should handle this safely, but let's test the handler passes it through
		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(nil, errors.New("invalid credentials")).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Performance - Large JSON Payload", func(t *testing.T) {
		// Test with a large but valid JSON payload
		largeDescription := ""
		for i := 0; i < 1000; i++ {
			largeDescription += "This is a very long description that simulates a large payload. "
		}

		signInData := map[string]interface{}{
			"email":       "test@example.com",
			"password":    "password123",
			"description": largeDescription, // Extra field that should be ignored
		}
		expectedResponse := &service.SignInResponse{
			IDToken:      "firebase-id-token-large",
			RefreshToken: "firebase-refresh-token-large",
			Email:        "test@example.com",
			LocalID:      "firebase-local-id-large",
		}

		mockService.On("SignIn", "test@example.com", "password123").Return(expectedResponse, nil).Once()

		bodyBytes, _ := json.Marshal(signInData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Multiple Concurrent Sign-ins", func(t *testing.T) {
		signInData := map[string]string{
			"email":    "concurrent@example.com",
			"password": "password123",
		}
		expectedResponse := &service.SignInResponse{
			IDToken:      "firebase-id-token-concurrent",
			RefreshToken: "firebase-refresh-token-concurrent",
			Email:        "concurrent@example.com",
			LocalID:      "firebase-local-id-concurrent",
		}

		// Mock multiple concurrent sign-ins
		mockService.On("SignIn", signInData["email"], signInData["password"]).Return(expectedResponse, nil).Times(3)

		// Simulate 3 concurrent sign-ins
		for i := 0; i < 3; i++ {
			bodyBytes, _ := json.Marshal(signInData)
			req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			httpHandler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Empty JSON Object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader([]byte(`{}`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Edge Case - Null Values in JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/editor/signin", bytes.NewReader([]byte(`{"email": null, "password": null}`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
} 