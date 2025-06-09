package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// MockAuthClient mocks the AuthClient interface
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) VerifyIDToken(ctx context.Context, token string) (*VerifiedToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VerifiedToken), args.Error(1)
}

// MockEditorRepository mocks the editor repository
type MockEditorRepository struct {
	mock.Mock
}

func (m *MockEditorRepository) GetEditorByFirebaseUID(ctx context.Context, firebaseUID string) (*models.Editor, error) {
	args := m.Called(ctx, firebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Editor), args.Error(1)
}

func (m *MockEditorRepository) InsertEditor(ctx context.Context, firebaseUID, email string) (*models.Editor, error) {
	args := m.Called(ctx, firebaseUID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Editor), args.Error(1)
}

func (m *MockEditorRepository) GetEditorByID(ctx context.Context, id string) (*models.Editor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Editor), args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		authHeader         string
		setupMocks         func(*MockAuthClient, *MockEditorRepository)
		expectedStatusCode int
		expectedError      string
		shouldSetContext   bool
	}{
		{
			name:       "successful authentication",
			authHeader: "Bearer valid_jwt_token",
			setupMocks: func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {
				verifiedToken := &VerifiedToken{UID: "firebase_user_123"}
				authClient.On("VerifyIDToken", mock.Anything, "valid_jwt_token").
					Return(verifiedToken, nil)

				editor := &models.Editor{
					ID:          "editor_456",
					FirebaseUID: "firebase_user_123",
					Email:       "test@example.com",
				}
				editorRepo.On("GetEditorByFirebaseUID", mock.Anything, "firebase_user_123").
					Return(editor, nil)
			},
			expectedStatusCode: http.StatusOK,
			shouldSetContext:   true,
		},
		{
			name:               "missing authorization header",
			authHeader:         "",
			setupMocks:         func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "Authorization header required",
		},
		{
			name:               "invalid authorization header format",
			authHeader:         "InvalidFormat token",
			setupMocks:         func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "Authorization header required",
		},
		{
			name:               "empty bearer token",
			authHeader:         "Bearer ",
			setupMocks:         func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "Bearer token cannot be empty",
		},
		{
			name:       "invalid JWT token",
			authHeader: "Bearer invalid_token",
			setupMocks: func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {
				authClient.On("VerifyIDToken", mock.Anything, "invalid_token").
					Return(nil, assert.AnError)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "Invalid or expired token",
		},
		{
			name:       "empty UID in verified token",
			authHeader: "Bearer valid_token_empty_uid",
			setupMocks: func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {
				verifiedToken := &VerifiedToken{UID: ""} // Empty UID
				authClient.On("VerifyIDToken", mock.Anything, "valid_token_empty_uid").
					Return(verifiedToken, nil)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "Invalid token",
		},
		{
			name:       "editor not found in database",
			authHeader: "Bearer valid_token_no_editor",
			setupMocks: func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {
				verifiedToken := &VerifiedToken{UID: "firebase_nonexistent"}
				authClient.On("VerifyIDToken", mock.Anything, "valid_token_no_editor").
					Return(verifiedToken, nil)

				editorRepo.On("GetEditorByFirebaseUID", mock.Anything, "firebase_nonexistent").
					Return(nil, apperrors.ErrEditorNotFound)
			},
			expectedStatusCode: http.StatusForbidden,
			expectedError:      "Editor not found",
		},
		{
			name:       "database error when fetching editor",
			authHeader: "Bearer valid_token_db_error",
			setupMocks: func(authClient *MockAuthClient, editorRepo *MockEditorRepository) {
				verifiedToken := &VerifiedToken{UID: "firebase_db_error"}
				authClient.On("VerifyIDToken", mock.Anything, "valid_token_db_error").
					Return(verifiedToken, nil)

				editorRepo.On("GetEditorByFirebaseUID", mock.Anything, "firebase_db_error").
					Return(nil, assert.AnError)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "Failed to retrieve editor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAuthClient := &MockAuthClient{}
			mockEditorRepo := &MockEditorRepository{}
			tt.setupMocks(mockAuthClient, mockEditorRepo)

			// Create a test handler that verifies context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.shouldSetContext {
					// Verify editor is in context
					editor, ok := GetEditorFromContext(r.Context())
					assert.True(t, ok, "Editor should be in context")
					assert.NotNil(t, editor, "Editor should not be nil")
					assert.Equal(t, "editor_456", editor.ID)

					// Verify editor ID is in context
					editorID := GetEditorIDFromContext(r.Context())
					assert.Equal(t, "editor_456", editorID)

					// Verify Firebase UID is in context
					firebaseUID := GetFirebaseUIDFromContext(r.Context())
					assert.Equal(t, "firebase_user_123", firebaseUID)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware
			middleware := AuthMiddleware(mockAuthClient, mockEditorRepo)
			handler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}

			// Verify mock expectations
			mockAuthClient.AssertExpectations(t)
			mockEditorRepo.AssertExpectations(t)
		})
	}
}

func TestGetEditorFromContext(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedEditor *models.Editor
		expectedOk     bool
	}{
		{
			name: "editor in context",
			setupContext: func() context.Context {
				editor := &models.Editor{
					ID:          "editor_123",
					FirebaseUID: "firebase_456",
					Email:       "test@example.com",
				}
				return context.WithValue(context.Background(), EditorContextKey, editor)
			},
			expectedEditor: &models.Editor{
				ID:          "editor_123",
				FirebaseUID: "firebase_456",
				Email:       "test@example.com",
			},
			expectedOk: true,
		},
		{
			name: "no editor in context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedEditor: nil,
			expectedOk:     false,
		},
		{
			name: "wrong type in context",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), EditorContextKey, "not an editor")
			},
			expectedEditor: nil,
			expectedOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			editor, ok := GetEditorFromContext(ctx)

			assert.Equal(t, tt.expectedOk, ok)
			if tt.expectedEditor != nil {
				assert.Equal(t, tt.expectedEditor.ID, editor.ID)
				assert.Equal(t, tt.expectedEditor.FirebaseUID, editor.FirebaseUID)
				assert.Equal(t, tt.expectedEditor.Email, editor.Email)
			} else {
				assert.Nil(t, editor)
			}
		})
	}
}

func TestGetEditorIDFromContext(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		expectedID string
	}{
		{
			name:       "editor ID in context",
			ctx:        context.WithValue(context.Background(), EditorIDContextKey, "editor_123"),
			expectedID: "editor_123",
		},
		{
			name:       "no editor ID in context",
			ctx:        context.Background(),
			expectedID: "",
		},
		{
			name:       "wrong type in context",
			ctx:        context.WithValue(context.Background(), EditorIDContextKey, 123),
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEditorIDFromContext(tt.ctx)
			assert.Equal(t, tt.expectedID, result)
		})
	}
} 