package post_handler_test

import (
	"bytes"
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

func TestUpdatePostHandler(t *testing.T) {
	// Store original and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	httpHandler := h.UpdatePostHandler(mockService)

	postID := uuid.New()
	editorFirebaseUID := "test-firebase-uid"

	t.Run("Success - Update Title Only", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		newTitle := "Updated Post Title"
		updateData := h.UpdatePostRequest{
			Title:   &newTitle,
			Content: nil,
		}
		expectedPost := &models.Post{
			ID:      postID,
			Title:   newTitle,
			Content: "Original content remains",
		}

		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, &newTitle, (*string)(nil)).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
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

	t.Run("Success - Update Content Only", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		newContent := "Updated post content with new information."
		updateData := h.UpdatePostRequest{
			Title:   nil,
			Content: &newContent,
		}
		expectedPost := &models.Post{
			ID:      postID,
			Title:   "Original title remains",
			Content: newContent,
		}

		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, (*string)(nil), &newContent).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resultPost models.Post
		err := json.Unmarshal(rr.Body.Bytes(), &resultPost)
		require.NoError(t, err)
		assert.Equal(t, expectedPost.Content, resultPost.Content)

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Update Both Title and Content", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		newTitle := "Completely Updated Title"
		newContent := "Completely updated content with all new information."
		updateData := h.UpdatePostRequest{
			Title:   &newTitle,
			Content: &newContent,
		}
		expectedPost := &models.Post{
			ID:      postID,
			Title:   newTitle,
			Content: newContent,
		}

		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, &newTitle, &newContent).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
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

		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader([]byte(`{}`)))
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

		req := httptest.NewRequest(http.MethodPut, "/api/posts/", nil)
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

		req := httptest.NewRequest(http.MethodPut, "/api/posts/invalid-uuid", nil)
		req.SetPathValue("postID", "invalid-uuid")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid post ID format")
	})

	t.Run("Error - Invalid JSON Body", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader([]byte(`{"title": "Test", content`)))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request body")
	})

	t.Run("Error - Empty Title if Provided", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		emptyTitle := ""
		updateData := h.UpdatePostRequest{
			Title:   &emptyTitle,
			Content: nil,
		}
		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post title, if provided, cannot be empty")
	})

	t.Run("Error - Empty Content if Provided", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		emptyContent := ""
		updateData := h.UpdatePostRequest{
			Title:   nil,
			Content: &emptyContent,
		}
		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post content, if provided, cannot be empty")
	})

	t.Run("Error - No Fields Provided for Update", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		updateData := h.UpdatePostRequest{
			Title:   nil,
			Content: nil,
		}
		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "At least one field (title or content) must be provided for update")
	})

	t.Run("Error - Service ErrPostNotFound", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		newTitle := "Updated Title"
		updateData := h.UpdatePostRequest{
			Title:   &newTitle,
			Content: nil,
		}
		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, &newTitle, (*string)(nil)).Return(nil, service.ErrPostNotFound).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
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

		newTitle := "Updated Title"
		updateData := h.UpdatePostRequest{
			Title:   &newTitle,
			Content: nil,
		}
		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, &newTitle, (*string)(nil)).Return(nil, service.ErrForbidden).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
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

		newTitle := "Updated Title"
		updateData := h.UpdatePostRequest{
			Title:   &newTitle,
			Content: nil,
		}
		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, &newTitle, (*string)(nil)).Return(nil, errors.New("database error")).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to update post")
		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Long Content Update", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		longContent := ""
		for i := 0; i < 500; i++ {
			longContent += "This is updated content that is very long to simulate real-world usage. "
		}

		updateData := h.UpdatePostRequest{
			Title:   nil,
			Content: &longContent,
		}
		expectedPost := &models.Post{
			ID:      postID,
			Title:   "Original title",
			Content: longContent,
		}

		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, (*string)(nil), &longContent).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Special Characters in Update", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		specialTitle := "Updated: áéíóú ñ ¿¡ 中文 🚀 & More!"
		specialContent := "Updated content with special chars: <script>alert('xss')</script> & SQL'; DROP TABLE posts; --"
		updateData := h.UpdatePostRequest{
			Title:   &specialTitle,
			Content: &specialContent,
		}
		expectedPost := &models.Post{
			ID:      postID,
			Title:   specialTitle,
			Content: specialContent,
		}

		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, &specialTitle, &specialContent).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(updateData)
		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader(bodyBytes))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Edge Case - Null Values in JSON", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		// Test explicit null values in JSON
		jsonBody := `{"title": null, "content": "Only content updated"}`
		expectedContent := "Only content updated"
		expectedPost := &models.Post{
			ID:      postID,
			Title:   "Original title",
			Content: expectedContent,
		}

		mockService.On("UpdatePost", mock.Anything, editorFirebaseUID, postID, (*string)(nil), &expectedContent).Return(expectedPost, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/api/posts/"+postID.String(), bytes.NewReader([]byte(jsonBody)))
		req.SetPathValue("postID", postID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})
} 