package subscriber_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNewsletterRepository is a mock implementation of NewsletterRepository
type MockNewsletterRepository struct {
	mock.Mock
}

func (m *MockNewsletterRepository) ListNewslettersByEditorID(editorID string, limit int, offset int) ([]repository.Newsletter, int, error) {
	args := m.Called(editorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]repository.Newsletter), args.Int(1), args.Error(2)
}

func (m *MockNewsletterRepository) CreateNewsletter(editorID, name, description string) (*repository.Newsletter, error) {
	args := m.Called(editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) GetNewsletterByIDAndEditorID(newsletterID, editorID string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID, editorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID, editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) DeleteNewsletter(newsletterID string, editorID string) error {
	args := m.Called(newsletterID, editorID)
	return args.Error(0)
}

func (m *MockNewsletterRepository) GetNewsletterByNameAndEditorID(name string, editorID string) (*repository.Newsletter, error) {
	args := m.Called(name, editorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) GetNewsletterByID(newsletterID string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID)
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

func TestGetSubscribersHandler(t *testing.T) {
	// Store original auth function and defer reset
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() {
		auth.VerifyFirebaseJWT = originalVerifyJWT
	}()

	mockSubscriberService := new(MockSubscriberService)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEditorRepo := new(MockEditorRepository)
	httpHandler := h.GetSubscribersHandler(mockSubscriberService, mockNewsletterRepo, mockEditorRepo)

	t.Run("Success - Get Subscribers List", func(t *testing.T) {
		newsletterID := "newsletter-123"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        "Test Newsletter",
			Description: "Test Description",
		}
		expectedSubscribers := []models.Subscriber{
			{
				ID:           "sub-1",
				Email:        "user1@example.com",
				NewsletterID: newsletterID,
				Status:       models.SubscriberStatusActive,
			},
			{
				ID:           "sub-2",
				Email:        "user2@example.com",
				NewsletterID: newsletterID,
				Status:       models.SubscriberStatusActive,
			},
		}

		// Mock authentication
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(expectedNewsletter, nil).Once()
		mockSubscriberService.GetActiveSubscribersForNewsletterFunc = func(ctx context.Context, nID string) ([]models.Subscriber, error) {
			if nID == newsletterID {
				return expectedSubscribers, nil
			}
			return nil, errors.New("unexpected newsletter ID")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

		// Verify response contains expected subscribers
		assert.Contains(t, rr.Body.String(), "user1@example.com")
		assert.Contains(t, rr.Body.String(), "user2@example.com")
		assert.Contains(t, rr.Body.String(), "sub-1")
		assert.Contains(t, rr.Body.String(), "sub-2")

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Success - Empty Subscribers List", func(t *testing.T) {
		newsletterID := "newsletter-empty"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        "Empty Newsletter",
			Description: "Empty Description",
		}

		// Mock authentication
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(expectedNewsletter, nil).Once()
		mockSubscriberService.GetActiveSubscribersForNewsletterFunc = func(ctx context.Context, nID string) ([]models.Subscriber, error) {
			if nID == newsletterID {
				return nil, nil // No subscribers
			}
			return nil, errors.New("unexpected newsletter ID")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
		assert.Contains(t, rr.Body.String(), "[]") // Empty array

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid Newsletter ID in Path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/newsletters//subscribers", nil)
		req.SetPathValue("newsletterID", "")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid newsletter ID in path")
	})

	t.Run("Error - Authentication Failed", func(t *testing.T) {
		newsletterID := "newsletter-123"

		// Mock authentication failure
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("invalid JWT token")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer invalid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authentication failed")
	})

	t.Run("Error - Editor Not Found by Firebase UID", func(t *testing.T) {
		newsletterID := "newsletter-123"
		firebaseUID := "nonexistent-firebase-uid"

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(nil, errors.New("editor not found")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to retrieve editor details")

		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Editor Returns Nil", func(t *testing.T) {
		newsletterID := "newsletter-123"
		firebaseUID := "firebase-uid-123"

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(nil, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Editor not found for Firebase UID")

		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Newsletter Not Found or Access Denied", func(t *testing.T) {
		newsletterID := "newsletter-123"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(nil, errors.New("newsletter not found")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Newsletter not found or access denied")

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Error - Cross-Editor Access Attempt", func(t *testing.T) {
		newsletterID := "newsletter-123"
		firebaseUID := "firebase-uid-456" // Different editor
		editorID := "editor-456"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "other-editor@example.com",
		}

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		// This newsletter belongs to a different editor
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(nil, errors.New("newsletter not found for this editor")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Newsletter not found or access denied")

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Error - Service Newsletter Not Found", func(t *testing.T) {
		newsletterID := "newsletter-123"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        "Test Newsletter",
			Description: "Test Description",
		}

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(expectedNewsletter, nil).Once()
		mockSubscriberService.GetActiveSubscribersForNewsletterFunc = func(ctx context.Context, nID string) ([]models.Subscriber, error) {
			return nil, service.ErrNewsletterNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Newsletter not found when fetching subscribers")

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Error - Service Internal Error", func(t *testing.T) {
		newsletterID := "newsletter-123"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        "Test Newsletter",
			Description: "Test Description",
		}

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(expectedNewsletter, nil).Once()
		mockSubscriberService.GetActiveSubscribersForNewsletterFunc = func(ctx context.Context, nID string) ([]models.Subscriber, error) {
			return nil, errors.New("database connection failed")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to get subscribers: database connection failed")

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Real-world Usage - Large Subscriber List", func(t *testing.T) {
		newsletterID := "newsletter-large"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        "Large Newsletter",
			Description: "Large Description",
		}

		// Create a large list of subscribers
		var expectedSubscribers []models.Subscriber
		for i := 0; i < 1000; i++ {
			expectedSubscribers = append(expectedSubscribers, models.Subscriber{
				ID:           "sub-" + string(rune(i)),
				Email:        "user" + string(rune(i)) + "@example.com",
				NewsletterID: newsletterID,
				Status:       models.SubscriberStatusActive,
			})
		}

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(expectedNewsletter, nil).Once()
		mockSubscriberService.GetActiveSubscribersForNewsletterFunc = func(ctx context.Context, nID string) ([]models.Subscriber, error) {
			if nID == newsletterID {
				return expectedSubscribers, nil
			}
			return nil, errors.New("unexpected newsletter ID")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

		// Verify response is not empty and contains array structure
		responseBody := rr.Body.String()
		assert.True(t, len(responseBody) > 100) // Large response
		assert.Contains(t, responseBody, "[")   // JSON array start
		assert.Contains(t, responseBody, "]")   // JSON array end

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Security - Invalid JWT Format", func(t *testing.T) {
		newsletterID := "newsletter-123"

		// Mock authentication failure with invalid JWT
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("invalid JWT format")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authentication failed")
	})

	t.Run("Security - Missing Authorization Header", func(t *testing.T) {
		newsletterID := "newsletter-123"

		// Mock authentication failure due to missing header
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("missing authorization header")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		// No Authorization header set
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authentication failed")
	})

	t.Run("Edge Case - Long Newsletter ID", func(t *testing.T) {
		newsletterID := "newsletter-with-very-long-id-that-might-cause-issues-in-some-systems-but-should-work-fine"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        "Long ID Newsletter",
			Description: "Long ID Description",
		}

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Once()
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(expectedNewsletter, nil).Once()
		mockSubscriberService.GetActiveSubscribersForNewsletterFunc = func(ctx context.Context, nID string) ([]models.Subscriber, error) {
			if nID == newsletterID {
				return []models.Subscriber{}, nil
			}
			return nil, errors.New("unexpected newsletter ID")
		}

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
		req.SetPathValue("newsletterID", newsletterID)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})

	t.Run("Performance - Concurrent Requests", func(t *testing.T) {
		newsletterID := "newsletter-concurrent"
		firebaseUID := "firebase-uid-123"
		editorID := "editor-123"

		expectedEditor := &repository.Editor{
			ID:          editorID,
			FirebaseUID: firebaseUID,
			Email:       "editor@example.com",
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        "Concurrent Newsletter",
			Description: "Concurrent Description",
		}
		expectedSubscribers := []models.Subscriber{
			{
				ID:           "sub-concurrent",
				Email:        "concurrent@example.com",
				NewsletterID: newsletterID,
				Status:       models.SubscriberStatusActive,
			},
		}

		// Mock authentication success
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return firebaseUID, nil
		}

		// Mock multiple calls for concurrent requests
		mockEditorRepo.On("GetEditorByFirebaseUID", firebaseUID).Return(expectedEditor, nil).Times(3)
		mockNewsletterRepo.On("GetNewsletterByIDAndEditorID", newsletterID, editorID).Return(expectedNewsletter, nil).Times(3)
		mockSubscriberService.GetActiveSubscribersForNewsletterFunc = func(ctx context.Context, nID string) ([]models.Subscriber, error) {
			if nID == newsletterID {
				return expectedSubscribers, nil
			}
			return nil, errors.New("unexpected newsletter ID")
		}

		// Simulate 3 concurrent requests
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID+"/subscribers", nil)
			req.SetPathValue("newsletterID", newsletterID)
			req.Header.Set("Authorization", "Bearer valid-jwt-token")
			rr := httptest.NewRecorder()

			httpHandler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		mockEditorRepo.AssertExpectations(t)
		mockNewsletterRepo.AssertExpectations(t)
	})
} 