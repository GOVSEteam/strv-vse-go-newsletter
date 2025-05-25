package post_handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockNewsletterService is a mock implementation of NewsletterServiceInterface
type MockNewsletterService struct {
	mock.Mock
}

func (m *MockNewsletterService) CreatePost(ctx context.Context, editorFirebaseUID string, newsletterID uuid.UUID, title string, content string) (*models.Post, error) {
	args := m.Called(ctx, editorFirebaseUID, newsletterID, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockNewsletterService) UpdatePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID, title *string, content *string) (*models.Post, error) {
	args := m.Called(ctx, editorFirebaseUID, postID, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockNewsletterService) DeletePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error {
	args := m.Called(ctx, editorFirebaseUID, postID)
	return args.Error(0)
}

func (m *MockNewsletterService) GetPostByID(ctx context.Context, postID uuid.UUID) (*models.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockNewsletterService) ListPostsByNewsletter(ctx context.Context, newsletterID uuid.UUID, limit int, offset int) ([]*models.Post, int, error) {
	args := m.Called(ctx, newsletterID, limit, offset)
	if args.Get(0) == nil {
		if args.Error(2) == nil {
			return []*models.Post{}, args.Int(1), nil
		}
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Int(1), args.Error(2)
}

func (m *MockNewsletterService) MarkPostAsPublished(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error {
	args := m.Called(ctx, editorFirebaseUID, postID)
	return args.Error(0)
}

func (m *MockNewsletterService) GetPostForPublishing(ctx context.Context, postID uuid.UUID, editorFirebaseUID string) (*models.Post, error) {
	args := m.Called(ctx, postID, editorFirebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

// Newsletter service methods (required by interface but not used in post handlers)
func (m *MockNewsletterService) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]repository.Newsletter, int, error) {
	args := m.Called(ctx, editorID, limit, offset)
	if args.Get(0) == nil {
		if args.Error(2) == nil {
			return []repository.Newsletter{}, args.Int(1), nil
		}
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]repository.Newsletter), args.Int(1), args.Error(2)
}

func (m *MockNewsletterService) CreateNewsletter(ctx context.Context, editorID, name, description string) (*repository.Newsletter, error) {
	args := m.Called(ctx, editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterService) UpdateNewsletter(ctx context.Context, newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error) {
	args := m.Called(ctx, newsletterID, editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterService) DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error {
	args := m.Called(ctx, newsletterID, editorID)
	return args.Error(0)
}

func (m *MockNewsletterService) GetNewsletterByID(ctx context.Context, newsletterID string) (*repository.Newsletter, error) {
	args := m.Called(ctx, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

// MockEditorRepository is a mock implementation of EditorRepository
type MockEditorRepository struct {
	mock.Mock
}

func (m *MockEditorRepository) InsertEditor(firebaseUID, email string) (*repository.Editor, error) {
	args := m.Called(firebaseUID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Editor), args.Error(1)
}

func (m *MockEditorRepository) GetEditorByFirebaseUID(firebaseUID string) (*repository.Editor, error) {
	args := m.Called(firebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Editor), args.Error(1)
}

func TestCreatePostHandler(t *testing.T) {
	// Store original and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	mockEditorRepo := new(MockEditorRepository)

	httpHandler := h.CreatePostHandler(mockService, mockEditorRepo)

	newsletterID := uuid.New()
	editorFirebaseUID := "test-firebase-uid"

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		postData := h.CreatePostRequest{
			Title:   "Test Post Title",
			Content: "This is the test post content.",
		}
		expectedPost := &models.Post{
			ID:           uuid.New(),
			NewsletterID: newsletterID,
			Title:        postData.Title,
			Content:      postData.Content,
		}

		mockService.On("CreatePost", mock.Anything, editorFirebaseUID, newsletterID, postData.Title, postData.Content).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resultPost models.Post
		err := json.Unmarshal(rr.Body.Bytes(), &resultPost)
		require.NoError(t, err)
		assert.Equal(t, expectedPost.Title, resultPost.Title)
		assert.Equal(t, expectedPost.Content, resultPost.Content)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID.String()+"/posts", nil)
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - Missing Newsletter ID in Path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters//posts", nil)
		req.SetPathValue("newsletterID", "")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Newsletter ID is required in path")
	})

	t.Run("Error - Invalid Newsletter ID Format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/invalid-uuid/posts", nil)
		req.SetPathValue("newsletterID", "invalid-uuid")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid newsletter ID format")
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt verification failed")
		}

		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader([]byte(`{}`)))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid or missing token")
	})

	t.Run("Error - Invalid JSON Body", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader([]byte(`{"title": "Test", content`)))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request body")
	})

	t.Run("Error - Empty Title", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		postData := h.CreatePostRequest{
			Title:   "",
			Content: "Valid content",
		}
		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post title cannot be empty")
	})

	t.Run("Error - Empty Content", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		postData := h.CreatePostRequest{
			Title:   "Valid title",
			Content: "",
		}
		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Post content cannot be empty")
	})

	t.Run("Error - Service ErrForbidden", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		postData := h.CreatePostRequest{
			Title:   "Test Title",
			Content: "Test Content",
		}
		mockService.On("CreatePost", mock.Anything, editorFirebaseUID, newsletterID, postData.Title, postData.Content).Return(nil, service.ErrForbidden).Once()

		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Forbidden")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service ErrServiceNewsletterNotFound", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		postData := h.CreatePostRequest{
			Title:   "Test Title",
			Content: "Test Content",
		}
		mockService.On("CreatePost", mock.Anything, editorFirebaseUID, newsletterID, postData.Title, postData.Content).Return(nil, service.ErrServiceNewsletterNotFound).Once()

		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Newsletter not found")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		postData := h.CreatePostRequest{
			Title:   "Test Title",
			Content: "Test Content",
		}
		mockService.On("CreatePost", mock.Anything, editorFirebaseUID, newsletterID, postData.Title, postData.Content).Return(nil, errors.New("database error")).Once()

		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to create post")
		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Long Content", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		longContent := ""
		for i := 0; i < 1000; i++ {
			longContent += "This is a very long post content that simulates real-world usage. "
		}

		postData := h.CreatePostRequest{
			Title:   "Long Content Post",
			Content: longContent,
		}
		expectedPost := &models.Post{
			ID:           uuid.New(),
			NewsletterID: newsletterID,
			Title:        postData.Title,
			Content:      postData.Content,
		}

		mockService.On("CreatePost", mock.Anything, editorFirebaseUID, newsletterID, postData.Title, postData.Content).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Special Characters", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return editorFirebaseUID, nil
		}

		postData := h.CreatePostRequest{
			Title:   "Post with Special Characters: áéíóú ñ ¿¡ 中文 🚀",
			Content: "Content with special chars: <script>alert('xss')</script> & SQL'; DROP TABLE posts; --",
		}
		expectedPost := &models.Post{
			ID:           uuid.New(),
			NewsletterID: newsletterID,
			Title:        postData.Title,
			Content:      postData.Content,
		}

		mockService.On("CreatePost", mock.Anything, editorFirebaseUID, newsletterID, postData.Title, postData.Content).Return(expectedPost, nil).Once()

		bodyBytes, _ := json.Marshal(postData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID.String()+"/posts", bytes.NewReader(bodyBytes))
		req.SetPathValue("newsletterID", newsletterID.String())
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})
} 