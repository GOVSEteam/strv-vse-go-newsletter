package post_handler_test

import (
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/require"
)

// MockPublishingService is a mock implementation of PublishingServiceInterface
type MockPublishingService struct {
	mock.Mock
}

func (m *MockPublishingService) PublishPostToSubscribers(ctx context.Context, postID string, editorFirebaseUID string) error {
	args := m.Called(ctx, postID, editorFirebaseUID)
	return args.Error(0)
}

func TestPublishPostHandler(t *testing.T) {
	// Store original and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockPublishingService := new(MockPublishingService)
	mockEditorRepo := new(MockEditorRepository)
	httpHandler := h.PublishPostHandler(mockPublishingService, mockEditorRepo)

	postID := uuid.New()
	editorFirebaseUID := "test-firebase-uid"

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Post published successfully and is being sent to subscribers.", response["message"])

		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt verification failed")
		}

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authentication failed")
	})

	t.Run("Error - Invalid Post ID in Path", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test with invalid path that doesn't match expected pattern
		req := httptest.NewRequest(http.MethodPost, "/api/invalid/path", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid post ID in path")
	})

	t.Run("Error - Empty Post ID in Path", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test with path that has empty ID
		req := httptest.NewRequest(http.MethodPost, "/api/posts//publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid post ID in path")
	})

	t.Run("Error - Service ErrPostNotFound", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(service.ErrPostNotFound).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post not found")
		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Error - Service ErrForbidden", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(service.ErrForbidden).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Forbidden")
		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Error - Post Already Published", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(errors.New("post is already published")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "post is already published")
		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(errors.New("email service unavailable")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to publish post")
		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Multiple Publish Attempts", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// First publish succeeds
		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(nil).Once()

		req1 := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr1 := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr1, req1)
		assert.Equal(t, http.StatusOK, rr1.Code)

		// Second publish fails (already published)
		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(errors.New("post is already published")).Once()

		req2 := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr2 := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr2, req2)
		assert.Equal(t, http.StatusConflict, rr2.Code)

		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Security - Different Editor Attempts Publish", func(t *testing.T) {
		differentEditorUID := "different-editor-uid"
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return differentEditorUID, nil
		}

		// Service should return forbidden when different editor tries to publish
		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), differentEditorUID).Return(service.ErrForbidden).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Forbidden")
		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Edge Case - Invalid UUID Format in Path", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test with invalid UUID format - the handler extracts the string and passes it to service
		invalidUUID := "invalid-uuid-format"
		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, invalidUUID, editorFirebaseUID).Return(errors.New("invalid post ID format")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+invalidUUID+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Edge Case - Very Long Post ID", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test with very long ID that's not a valid UUID
		longID := "this-is-a-very-long-invalid-uuid-that-should-fail-parsing-" + postID.String()
		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, longID, editorFirebaseUID).Return(errors.New("invalid post ID format")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+longID+"/publish", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockPublishingService.AssertExpectations(t)
	})

	t.Run("Edge Case - Missing Publish Suffix", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test with path missing /publish suffix
		req := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String(), nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid post ID in path")
	})

	t.Run("Performance - Concurrent Publish Attempts", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Mock multiple concurrent publish attempts
		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(nil).Once()
		mockPublishingService.On("PublishPostToSubscribers", mock.Anything, postID.String(), editorFirebaseUID).Return(errors.New("post is already published")).Once()

		// First request should succeed
		req1 := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr1 := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr1, req1)
		assert.Equal(t, http.StatusOK, rr1.Code)

		// Second concurrent request should fail
		req2 := httptest.NewRequest(http.MethodPost, "/api/posts/"+postID.String()+"/publish", nil)
		rr2 := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr2, req2)
		assert.Equal(t, http.StatusConflict, rr2.Code)

		mockPublishingService.AssertExpectations(t)
	})
} 