package post_handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListPostsByNewsletterHandler(t *testing.T) {
	// Store original and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	httpHandler := h.ListPostsByNewsletterHandler(mockService)

	newsletterID := uuid.New()
	editorFirebaseUID := "test-firebase-uid"

	t.Run("Success - Default Pagination", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		expectedPosts := []*models.Post{
			{
				ID:           uuid.New(),
				NewsletterID: newsletterID,
				Title:        "First Post",
				Content:      "Content of first post",
			},
			{
				ID:           uuid.New(),
				NewsletterID: newsletterID,
				Title:        "Second Post",
				Content:      "Content of second post",
			},
		}
		expectedTotal := 25

		mockService.On("ListPostsByNewsletter", mock.Anything, newsletterID, h.DefaultPostLimit, h.DefaultPostOffset).Return(expectedPosts, expectedTotal, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response h.PaginatedPostsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedTotal, response.Total)
		assert.Equal(t, h.DefaultPostLimit, response.Limit)
		assert.Equal(t, h.DefaultPostOffset, response.Offset)
		assert.Len(t, response.Data, len(expectedPosts))
		assert.Equal(t, expectedPosts[0].Title, response.Data[0].Title)

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Custom Pagination", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		customLimit := 5
		customOffset := 10
		expectedPosts := []*models.Post{
			{
				ID:           uuid.New(),
				NewsletterID: newsletterID,
				Title:        "Post from page 3",
				Content:      "Content from page 3",
			},
		}
		expectedTotal := 50

		mockService.On("ListPostsByNewsletter", mock.Anything, newsletterID, customLimit, customOffset).Return(expectedPosts, expectedTotal, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?limit=5&offset=10", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response h.PaginatedPostsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedTotal, response.Total)
		assert.Equal(t, customLimit, response.Limit)
		assert.Equal(t, customOffset, response.Offset)
		assert.Len(t, response.Data, len(expectedPosts))

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Empty Results", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		expectedPosts := []*models.Post{}
		expectedTotal := 0

		mockService.On("ListPostsByNewsletter", mock.Anything, newsletterID, h.DefaultPostLimit, h.DefaultPostOffset).Return(expectedPosts, expectedTotal, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response h.PaginatedPostsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 0, response.Total)
		assert.Len(t, response.Data, 0)

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Max Limit Enforcement", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Request limit higher than max, should be capped
		expectedPosts := []*models.Post{}
		expectedTotal := 200

		// Should use MaxPostLimit instead of requested limit
		mockService.On("ListPostsByNewsletter", mock.Anything, newsletterID, h.MaxPostLimit, h.DefaultPostOffset).Return(expectedPosts, expectedTotal, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?limit=150", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response h.PaginatedPostsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, h.MaxPostLimit, response.Limit) // Should be capped

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt verification failed")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid or missing token")
	})

	t.Run("Error - Missing Newsletter ID in Path", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters//posts", nil)
		req.SetPathValue("newsletterID", "")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Newsletter ID is required in path")
	})

	t.Run("Error - Invalid Newsletter ID Format", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/invalid-uuid/posts", nil)
		req.SetPathValue("newsletterID", "invalid-uuid")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid newsletter ID format")
	})

	t.Run("Error - Invalid Limit Parameter", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?limit=invalid", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid limit parameter")
	})

	t.Run("Error - Negative Limit Parameter", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?limit=-5", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid limit parameter")
	})

	t.Run("Error - Invalid Offset Parameter", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?offset=invalid", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid offset parameter")
	})

	t.Run("Error - Negative Offset Parameter", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?offset=-10", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid offset parameter")
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		mockService.On("ListPostsByNewsletter", mock.Anything, newsletterID, h.DefaultPostLimit, h.DefaultPostOffset).Return(nil, 0, errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to list posts")
		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Large Dataset Pagination", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Simulate pagination through a large dataset
		limit := 20
		offset := 100
		expectedTotal := 1000

		// Generate posts for page 6 (offset 100, limit 20)
		expectedPosts := make([]*models.Post, limit)
		for i := 0; i < limit; i++ {
			expectedPosts[i] = &models.Post{
				ID:           uuid.New(),
				NewsletterID: newsletterID,
				Title:        "Post " + string(rune(offset+i+1)),
				Content:      "Content for post number " + string(rune(offset+i+1)),
			}
		}

		mockService.On("ListPostsByNewsletter", mock.Anything, newsletterID, limit, offset).Return(expectedPosts, expectedTotal, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?limit=20&offset=100", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response h.PaginatedPostsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedTotal, response.Total)
		assert.Equal(t, limit, response.Limit)
		assert.Equal(t, offset, response.Offset)
		assert.Len(t, response.Data, limit)

		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Zero Limit", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?limit=0", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid limit parameter")
	})

	t.Run("Edge Case - Very Large Offset", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		largeOffset := 999999
		expectedPosts := []*models.Post{} // No posts at this offset
		expectedTotal := 100              // Total posts in newsletter

		mockService.On("ListPostsByNewsletter", mock.Anything, newsletterID, h.DefaultPostLimit, largeOffset).Return(expectedPosts, expectedTotal, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts?offset=999999", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response h.PaginatedPostsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedTotal, response.Total)
		assert.Equal(t, largeOffset, response.Offset)
		assert.Len(t, response.Data, 0) // No posts at this offset

		mockService.AssertExpectations(t)
	})
} 