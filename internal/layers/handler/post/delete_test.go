package post_handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeletePostHandler(t *testing.T) {
	// Store original and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	httpHandler := h.DeletePostHandler(mockService)

	postID := uuid.New()
	editorFirebaseUID := "test-firebase-uid"

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockService.On("DeletePost", mock.Anything, editorFirebaseUID, postID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, rr.Body.String()) // No content expected for 204

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt verification failed")
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid or missing token")
	})

	t.Run("Error - Missing Post ID in Path", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/", nil)
		req.SetPathValue("postID", "")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post ID is required in path")
	})

	t.Run("Error - Invalid Post ID Format", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/invalid-uuid", nil)
		req.SetPathValue("postID", "invalid-uuid")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid post ID format")
	})

	t.Run("Error - Service ErrPostNotFound", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockService.On("DeletePost", mock.Anything, editorFirebaseUID, postID).Return(service.ErrPostNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post not found")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service ErrForbidden", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockService.On("DeletePost", mock.Anything, editorFirebaseUID, postID).Return(service.ErrForbidden).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Forbidden")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockService.On("DeletePost", mock.Anything, editorFirebaseUID, postID).Return(errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to delete post")
		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Multiple Delete Attempts", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// First delete succeeds
		mockService.On("DeletePost", mock.Anything, editorFirebaseUID, postID).Return(nil).Once()

		req1 := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req1.SetPathValue("postID", postID.String())
		rr1 := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr1, req1)
		assert.Equal(t, http.StatusNoContent, rr1.Code)

		// Second delete fails (post already deleted)
		mockService.On("DeletePost", mock.Anything, editorFirebaseUID, postID).Return(service.ErrPostNotFound).Once()

		req2 := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req2.SetPathValue("postID", postID.String())
		rr2 := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr2, req2)
		assert.Equal(t, http.StatusNotFound, rr2.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("Security - Different Editor Attempts Delete", func(t *testing.T) {
		differentEditorUID := "different-editor-uid"
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return differentEditorUID, nil
		}

		// Service should return forbidden when different editor tries to delete
		mockService.On("DeletePost", mock.Anything, differentEditorUID, postID).Return(service.ErrForbidden).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Forbidden")
		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Very Long UUID", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test with malformed UUID that's very long
		longInvalidUUID := "this-is-a-very-long-invalid-uuid-that-should-fail-parsing-" + postID.String()
		req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+longInvalidUUID, nil)
		req.SetPathValue("postID", longInvalidUUID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid post ID format")
	})

	t.Run("Edge Case - Empty String UUID", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/api/posts/", nil)
		req.SetPathValue("postID", "")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post ID is required in path")
	})
} 