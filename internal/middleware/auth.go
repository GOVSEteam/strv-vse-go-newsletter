package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// EditorRepository defines the interface for editor data access needed by the auth middleware.
// This is a placeholder until the actual repository is refactored.
type EditorRepository interface {
	GetEditorByFirebaseUID(ctx context.Context, firebaseUID string) (*models.Editor, error)
}

// contextKey is a type used for context keys to avoid collisions.
type contextKey string

const (
	// EditorIDContextKey is the key for storing the editor ID in the context.
	EditorIDContextKey contextKey = "editorID"
	// FirebaseUIDContextKey is the key for storing the Firebase UID in the context.
	FirebaseUIDContextKey contextKey = "firebaseUID"
)

// AuthMiddleware creates a new authentication middleware.
// It verifies the Firebase JWT from the Authorization header, retrieves the editor,
// and injects editor ID and Firebase UID into the request context.
func AuthMiddleware(authClient *auth.Client, editorRepo EditorRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Use a helper that writes a JSON error response based on apperrors
				// For now, we will write a plain text error
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, "Bearer ")
			if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			token, err := authClient.VerifyIDToken(ctx, tokenString)
			if err != nil {
				statusCode := apperrors.ErrorToHTTPStatus(apperrors.ErrUnauthorized) // Default to Unauthorized
				// Firebase often returns specific error types that can be checked.
				// For example, if token is expired, revoked, etc.
				// For now, a generic unauthorized error is used.
				http.Error(w, fmt.Sprintf("Invalid or expired token: %v", err), statusCode)
				return
			}

			firebaseUID := token.UID
			editor, err := editorRepo.GetEditorByFirebaseUID(ctx, firebaseUID)
			if err != nil {
				if apperrors.IsNotFound(err) {
					http.Error(w, "Editor not found for the provided token", http.StatusForbidden)
				} else {
					// Log the error for internal tracking
					// log.Printf("Error getting editor by Firebase UID %s: %v", firebaseUID, err)
					http.Error(w, "Failed to retrieve editor information", http.StatusInternalServerError)
				}
				return
			}

			if editor == nil { // Double check in case GetEditorByFirebaseUID returns nil, nil
				http.Error(w, "Editor not found (nil editor returned)", http.StatusForbidden)
				return
			}

			// Store editor ID and Firebase UID in context
			ctx = context.WithValue(ctx, EditorIDContextKey, editor.ID)
			ctx = context.WithValue(ctx, FirebaseUIDContextKey, firebaseUID)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// GetEditorIDFromContext retrieves the editor ID from the request context.
// Returns an empty string if not found.
func GetEditorIDFromContext(ctx context.Context) string {
	if editorID, ok := ctx.Value(EditorIDContextKey).(string); ok {
		return editorID
	}
	return ""
}

// GetFirebaseUIDFromContext retrieves the Firebase UID from the request context.
// Returns an empty string if not found.
func GetFirebaseUIDFromContext(ctx context.Context) string {
	if firebaseUID, ok := ctx.Value(FirebaseUIDContextKey).(string); ok {
		return firebaseUID
	}
	return ""
} 