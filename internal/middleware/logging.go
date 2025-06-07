package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// contextKey is an unexported type for context keys to prevent collisions.
// This type is shared across all middleware in this package.
type contextKey string

// requestIDContextKey is the key for storing the request ID in the context.
const requestIDContextKey contextKey = "requestID"

// GetRequestIDFromContext retrieves the request ID from the context.
// Returns an empty string if not found.
func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return reqID
	}
	return ""
}

// responseWriterInterceptor is a wrapper around http.ResponseWriter to capture status code.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriterInterceptor creates a new responseWriterInterceptor.
func NewResponseWriterInterceptor(w http.ResponseWriter) *responseWriterInterceptor {
	// Default to 200 OK if WriteHeader is not called
	return &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader captures the status code.
func (rwi *responseWriterInterceptor) WriteHeader(statusCode int) {
	rwi.statusCode = statusCode
	rwi.ResponseWriter.WriteHeader(statusCode)
}

// LoggingMiddleware creates a new HTTP middleware for structured request logging.
func LoggingMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.NewString()
			ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
			r = r.WithContext(ctx)

			startTime := time.Now()
			rwi := NewResponseWriterInterceptor(w)
			next.ServeHTTP(rwi, r)
			duration := time.Since(startTime)

			logFields := []interface{}{
				"requestID", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rwi.statusCode,
				"duration", duration.String(),
				"remoteAddr", r.RemoteAddr,
			}

			if rwi.statusCode >= http.StatusInternalServerError {
				logger.Errorw("Request completed with server error", logFields...)
			} else if rwi.statusCode >= http.StatusBadRequest && rwi.statusCode < http.StatusInternalServerError {
				logger.Warnw("Request completed with client error", logFields...)
			} else {
				logger.Infow("Request completed", logFields...)
			}
		})
	}
}
