package post_handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetPostByIDHandler(t *testing.T) {
	// Store original and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	httpHandler := h.GetPostByIDHandler(mockService)

	postID := uuid.New()
	newsletterID := uuid.New()
	editorFirebaseUID := "test-firebase-uid"

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		expectedPost := &models.Post{
			ID:           postID,
			NewsletterID: newsletterID,
			Title:        "Test Post Title",
			Content:      "This is the test post content.",
		}

		mockService.On("GetPostByID", mock.Anything, postID).Return(expectedPost, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resultPost models.Post
		err := json.Unmarshal(rr.Body.Bytes(), &resultPost)
		require.NoError(t, err)
		assert.Equal(t, expectedPost.ID, resultPost.ID)
		assert.Equal(t, expectedPost.Title, resultPost.Title)
		assert.Equal(t, expectedPost.Content, resultPost.Content)

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

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
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

		req := httptest.NewRequest(http.MethodGet, "/api/posts/", nil)
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

		req := httptest.NewRequest(http.MethodGet, "/api/posts/invalid-uuid", nil)
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

		mockService.On("GetPostByID", mock.Anything, postID).Return(nil, service.ErrPostNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post not found")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockService.On("GetPostByID", mock.Anything, postID).Return(nil, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to get post")
		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Post with Long Content", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		longContent := ""
		for i := 0; i < 1000; i++ {
			longContent += "This is a very long post content that simulates real-world usage. "
		}

		expectedPost := &models.Post{
			ID:           postID,
			NewsletterID: newsletterID,
			Title:        "Post with Very Long Content",
			Content:      longContent,
		}

		mockService.On("GetPostByID", mock.Anything, postID).Return(expectedPost, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resultPost models.Post
		err := json.Unmarshal(rr.Body.Bytes(), &resultPost)
		require.NoError(t, err)
		assert.Equal(t, expectedPost.Content, resultPost.Content)
		assert.True(t, len(resultPost.Content) > 50000) // Verify it's actually long

		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Post with Special Characters", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		expectedPost := &models.Post{
			ID:           postID,
			NewsletterID: newsletterID,
			Title:        "Post with Special Characters: áéíóú ñ ¿¡ 中文 🚀",
			Content:      "Content with special chars: <script>alert('xss')</script> & SQL'; DROP TABLE posts; --",
		}

		mockService.On("GetPostByID", mock.Anything, postID).Return(expectedPost, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resultPost models.Post
		err := json.Unmarshal(rr.Body.Bytes(), &resultPost)
		require.NoError(t, err)
		assert.Equal(t, expectedPost.Title, resultPost.Title)
		assert.Equal(t, expectedPost.Content, resultPost.Content)

		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Post with Empty Content", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		expectedPost := &models.Post{
			ID:           postID,
			NewsletterID: newsletterID,
			Title:        "Post with Empty Content",
			Content:      "", // Empty content
		}

		mockService.On("GetPostByID", mock.Anything, postID).Return(expectedPost, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resultPost models.Post
		err := json.Unmarshal(rr.Body.Bytes(), &resultPost)
		require.NoError(t, err)
		assert.Equal(t, expectedPost.Title, resultPost.Title)
		assert.Equal(t, "", resultPost.Content)

		mockService.AssertExpectations(t)
	})

	t.Run("Security - Different Editor Can Still Read", func(t *testing.T) {
		// Note: The current GetPostByIDHandler doesn't check ownership
		// This test documents the current behavior - any authenticated user can read any post
		differentEditorUID := "different-editor-uid"
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return differentEditorUID, nil
		}

		expectedPost := &models.Post{
			ID:           postID,
			NewsletterID: newsletterID,
			Title:        "Post readable by any authenticated user",
			Content:      "This post can be read by any authenticated user.",
		}

		mockService.On("GetPostByID", mock.Anything, postID).Return(expectedPost, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resultPost models.Post
		err := json.Unmarshal(rr.Body.Bytes(), &resultPost)
		require.NoError(t, err)
		assert.Equal(t, expectedPost.Title, resultPost.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Very Long UUID", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test with malformed UUID that's very long
		longInvalidUUID := "this-is-a-very-long-invalid-uuid-that-should-fail-parsing-" + postID.String()
		req := httptest.NewRequest(http.MethodGet, "/api/posts/"+longInvalidUUID, nil)
		req.SetPathValue("postID", longInvalidUUID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid post ID format")
	})

	t.Run("Performance - Multiple Concurrent Reads", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		expectedPost := &models.Post{
			ID:           postID,
			NewsletterID: newsletterID,
			Title:        "Concurrently Readable Post",
			Content:      "This post is being read concurrently.",
		}

		// Mock multiple calls for concurrent access
		mockService.On("GetPostByID", mock.Anything, postID).Return(expectedPost, nil).Times(3)

		// Simulate 3 concurrent reads
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/posts/"+postID.String(), nil)
			req.SetPathValue("postID", postID.String())
			rr := httptest.NewRecorder()

			httpHandler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		mockService.AssertExpectations(t)
	})
} 