package middleware

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// VerifiedToken represents a verified authentication token (abstracted from Firebase).
type VerifiedToken struct {
	UID string
}

// AuthClient defines the interface for authentication operations (provider-agnostic).
type AuthClient interface {
	VerifyIDToken(ctx context.Context, token string) (*VerifiedToken, error)
}

const (
	EditorContextKey      = "editor"
	EditorIDContextKey    = "editorID"
	FirebaseUIDContextKey = "firebaseUID"
)

// AuthMiddleware creates authentication middleware for Firebase JWT tokens.
func AuthMiddleware(authClient AuthClient, editorRepo repository.EditorRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Bearer token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			token := strings.TrimSpace(authHeader[7:])
			if token == "" {
				http.Error(w, "Bearer token cannot be empty", http.StatusUnauthorized)
				return
			}

			// Verify Firebase ID token
			idToken, err := authClient.VerifyIDToken(r.Context(), token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			if idToken.UID == "" {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Get editor from database
			editor, err := editorRepo.GetEditorByFirebaseUID(r.Context(), idToken.UID)
			if err != nil {
				if apperrors.IsNotFound(err) {
					http.Error(w, "Editor not found", http.StatusForbidden)
				} else {
					http.Error(w, "Failed to retrieve editor", http.StatusInternalServerError)
				}
				return
			}

			// Store in context
			ctx := context.WithValue(r.Context(), EditorContextKey, editor)
			ctx = context.WithValue(ctx, EditorIDContextKey, editor.ID)
			ctx = context.WithValue(ctx, FirebaseUIDContextKey, idToken.UID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetEditorFromContext retrieves the editor from context.
func GetEditorFromContext(ctx context.Context) (*models.Editor, bool) {
	if editor, ok := ctx.Value(EditorContextKey).(*models.Editor); ok {
		return editor, true
	}
	return nil, false
}

// GetEditorIDFromContext retrieves the editor ID from context.
func GetEditorIDFromContext(ctx context.Context) string {
	if editorID, ok := ctx.Value(EditorIDContextKey).(string); ok {
		return editorID
	}
	return ""
}

// GetFirebaseUIDFromContext retrieves the Firebase UID from context.
func GetFirebaseUIDFromContext(ctx context.Context) string {
	if firebaseUID, ok := ctx.Value(FirebaseUIDContextKey).(string); ok {
		return firebaseUID
	}
	return ""
}

// FirebaseAuthAdapter adapts Firebase auth.Client to AuthClient interface.
// This adapter converts Firebase-specific types to domain types for the middleware.
type FirebaseAuthAdapter struct {
	client *auth.Client
}

// NewFirebaseAuthAdapter creates a new Firebase auth adapter.
func NewFirebaseAuthAdapter(client *auth.Client) AuthClient {
	return &FirebaseAuthAdapter{client: client}
}

// VerifyIDToken verifies a Firebase ID token and returns a domain VerifiedToken.
func (a *FirebaseAuthAdapter) VerifyIDToken(ctx context.Context, token string) (*VerifiedToken, error) {
	firebaseToken, err := a.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return &VerifiedToken{UID: firebaseToken.UID}, nil
}

