package editor_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEditorService is a mock implementation of EditorService
type MockEditorService struct {
	mock.Mock
}

func (m *MockEditorService) SignUp(email, password string) (*repository.Editor, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Editor), args.Error(1)
}

func (m *MockEditorService) SignIn(email, password string) (*service.SignInResponse, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SignInResponse), args.Error(1)
}

func TestEditorSignUpHandler(t *testing.T) {
	mockService := new(MockEditorService)
	httpHandler := h.EditorSignUpHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		expectedEditor := &repository.Editor{
			ID:          "editor-id-123",
			FirebaseUID: "firebase-uid-123",
			Email:       "test@example.com",
		}

		mockService.On("SignUp", signUpData["email"], signUpData["password"]).Return(expectedEditor, nil).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedEditor.ID, response["editor_id"])
		assert.Equal(t, expectedEditor.Email, response["email"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/editor/signup", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - Invalid JSON Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader([]byte(`{"email": "test@example.com", password`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Empty Email", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "",
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Empty Password", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "test@example.com",
			"password": "",
		}
		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Missing Email Field", func(t *testing.T) {
		signUpData := map[string]string{
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Missing Password Field", func(t *testing.T) {
		signUpData := map[string]string{
			"email": "test@example.com",
		}
		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Service Email Already Exists", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "existing@example.com",
			"password": "password123",
		}

		mockService.On("SignUp", signUpData["email"], signUpData["password"]).Return(nil, errors.New("email already exists")).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "email already exists", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Firebase Error", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		mockService.On("SignUp", signUpData["email"], signUpData["password"]).Return(nil, errors.New("firebase auth error")).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "firebase auth error", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Database Error", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		mockService.On("SignUp", signUpData["email"], signUpData["password"]).Return(nil, errors.New("database connection failed")).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "database connection failed", response["error"])

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
				signUpData := map[string]string{
					"email":    tc.email,
					"password": "password123",
				}
				expectedEditor := &repository.Editor{
					ID:          "editor-id-" + tc.name,
					FirebaseUID: "firebase-uid-" + tc.name,
					Email:       tc.email,
				}

				mockService.On("SignUp", tc.email, "password123").Return(expectedEditor, nil).Once()

				bodyBytes, _ := json.Marshal(signUpData)
				req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
				rr := httptest.NewRecorder()

				httpHandler.ServeHTTP(rr, req)

				assert.Equal(t, http.StatusCreated, rr.Code)
				mockService.AssertExpectations(t)
			})
		}
	})

	t.Run("Real-world Usage - Password Complexity", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "test@example.com",
			"password": "ComplexP@ssw0rd!123",
		}
		expectedEditor := &repository.Editor{
			ID:          "editor-id-complex",
			FirebaseUID: "firebase-uid-complex",
			Email:       "test@example.com",
		}

		mockService.On("SignUp", signUpData["email"], signUpData["password"]).Return(expectedEditor, nil).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Very Long Email", func(t *testing.T) {
		longEmail := "very.long.email.address.that.might.cause.issues@very.long.domain.name.example.com"
		signUpData := map[string]string{
			"email":    longEmail,
			"password": "password123",
		}
		expectedEditor := &repository.Editor{
			ID:          "editor-id-long",
			FirebaseUID: "firebase-uid-long",
			Email:       longEmail,
		}

		mockService.On("SignUp", longEmail, "password123").Return(expectedEditor, nil).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Special Characters in Password", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "test@example.com",
			"password": "P@$$w0rd!#$%^&*()_+-=[]{}|;:,.<>?",
		}
		expectedEditor := &repository.Editor{
			ID:          "editor-id-special",
			FirebaseUID: "firebase-uid-special",
			Email:       "test@example.com",
		}

		mockService.On("SignUp", signUpData["email"], signUpData["password"]).Return(expectedEditor, nil).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Security - SQL Injection Attempt in Email", func(t *testing.T) {
		signUpData := map[string]string{
			"email":    "test'; DROP TABLE editors; --@example.com",
			"password": "password123",
		}

		// Service should handle this safely, but let's test the handler passes it through
		mockService.On("SignUp", signUpData["email"], signUpData["password"]).Return(nil, errors.New("invalid email format")).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Performance - Large JSON Payload", func(t *testing.T) {
		// Test with a large but valid JSON payload
		largeDescription := ""
		for i := 0; i < 1000; i++ {
			largeDescription += "This is a very long description that simulates a large payload. "
		}

		signUpData := map[string]interface{}{
			"email":       "test@example.com",
			"password":    "password123",
			"description": largeDescription, // Extra field that should be ignored
		}
		expectedEditor := &repository.Editor{
			ID:          "editor-id-large",
			FirebaseUID: "firebase-uid-large",
			Email:       "test@example.com",
		}

		mockService.On("SignUp", "test@example.com", "password123").Return(expectedEditor, nil).Once()

		bodyBytes, _ := json.Marshal(signUpData)
		req := httptest.NewRequest(http.MethodPost, "/editor/signup", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})
} 