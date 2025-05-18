package newsletter_test

import (
	"bytes"
	"context" // Added context
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter" // Alias for handler package
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // Added for models.Post
	"github.com/google/uuid"                                      // Added for uuid.UUID
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

// MockNewsletterService is a mock implementation of NewsletterServiceInterface
type MockNewsletterService struct {
	mock.Mock
}

// --- Newsletter Methods ---
func (m *MockNewsletterService) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]repository.Newsletter, int, error) {
	args := m.Called(ctx, editorID, limit, offset)
	if args.Get(0) == nil {
		// Handle case where no newsletters are returned but no error (e.g., empty list)
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

// --- Post Methods (to satisfy NewsletterServiceInterface) ---
func (m *MockNewsletterService) CreatePost(ctx context.Context, editorFirebaseUID string, newsletterID uuid.UUID, title string, content string) (*models.Post, error) {
	args := m.Called(ctx, editorFirebaseUID, newsletterID, title, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
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

func (m *MockNewsletterService) MarkPostAsPublished(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error {
	args := m.Called(ctx, editorFirebaseUID, postID)
	return args.Error(0)
}

func TestCreateHandler(t *testing.T) {
	// Store original and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	mockEditorRepo := new(MockEditorRepository)

	httpHandler := h.CreateHandler(mockService, mockEditorRepo)

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		newsletterData := h.CreateNewsletterRequest{Name: "Test Newsletter", Description: "A test desc"}
		expectedNewsletter := &repository.Newsletter{
			ID:          "nl-123",
			EditorID:    "editor-123",
			Name:        newsletterData.Name,
			Description: newsletterData.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockService.On("CreateNewsletter", mock.Anything, "editor-123", newsletterData.Name, newsletterData.Description).Return(expectedNewsletter, nil).Once()

		bodyBytes, _ := json.Marshal(newsletterData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resultNewsletter repository.Newsletter
		json.Unmarshal(rr.Body.Bytes(), &resultNewsletter)
		assert.Equal(t, expectedNewsletter.Name, resultNewsletter.Name) // Compare relevant fields
		assert.Equal(t, expectedNewsletter.Description, resultNewsletter.Description)

		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/newsletters", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt error")
		}
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", bytes.NewReader([]byte(`{}`))) // Empty body fine for this check
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - Editor Not Found", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(nil, sql.ErrNoRows).Once()

		reqBody := h.CreateNewsletterRequest{Name: "Test", Description: "Test"}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", bytes.NewReader(bodyBytes))
		// Add context to request if needed by handlers, though ServeHTTP usually handles it.
		// req = req.WithContext(context.Background()) 
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid JSON Body", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { // Need to mock auth to reach body parsing
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", bytes.NewReader([]byte(`{"name": "Test", desc`)))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockEditorRepo.AssertExpectations(t) // Editor repo was called before body parsing failure
	})

	t.Run("Error - Empty Newsletter Name", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		reqBody := h.CreateNewsletterRequest{Name: "", Description: "Test"}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Service ErrNewsletterNameTaken", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		newsletterData := h.CreateNewsletterRequest{Name: "Taken Name", Description: "A test desc"}
		mockService.On("CreateNewsletter", mock.Anything, "editor-123", newsletterData.Name, newsletterData.Description).Return(nil, service.ErrNewsletterNameTaken).Once()

		bodyBytes, _ := json.Marshal(newsletterData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		newsletterData := h.CreateNewsletterRequest{Name: "Good Name", Description: "A test desc"}
		mockService.On("CreateNewsletter", mock.Anything, "editor-123", newsletterData.Name, newsletterData.Description).Return(nil, errors.New("some db error")).Once()

		bodyBytes, _ := json.Marshal(newsletterData)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})
}
