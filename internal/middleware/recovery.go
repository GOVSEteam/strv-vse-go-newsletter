package middleware

import (
	"net/http"
	"runtime"

	"go.uber.org/zap"
)

// RecoveryMiddleware creates a panic recovery middleware that:
// - Captures and logs panics with stack traces
// - Returns proper HTTP 500 responses
// - Ensures the application doesn't crash on panics
func RecoveryMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Capture stack trace
					stackBuf := make([]byte, 2048)
					stackSize := runtime.Stack(stackBuf, false)
					stackTrace := string(stackBuf[:stackSize])

					// Get request ID from context if available
					requestID := GetRequestIDFromContext(r.Context())

					// Log the panic with essential information
					logger.Errorw("Panic recovered",
						"requestID", requestID,
						"method", r.Method,
						"path", r.URL.Path,
						"panic", rec,
						"stackTrace", stackTrace,
					)

					// Return simple error response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"Internal server error"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
