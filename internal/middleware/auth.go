package middleware

import (
	"context"
	"net/http"
	"strings"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// Use the centralized FirebaseAuthClient interface from the service package
type AuthClient = service.FirebaseAuthClient

// EditorRepository defines the interface for editor data access.
type EditorRepository interface {
	GetEditorByFirebaseUID(ctx context.Context, firebaseUID string) (*models.Editor, error)
}

const (
	EditorContextKey      = "editor"
	EditorIDContextKey    = "editorID"
	FirebaseUIDContextKey = "firebaseUID"
)

// AuthMiddleware creates authentication middleware for Firebase JWT tokens.
func AuthMiddleware(authClient AuthClient, editorRepo EditorRepository) func(http.Handler) http.Handler {
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

