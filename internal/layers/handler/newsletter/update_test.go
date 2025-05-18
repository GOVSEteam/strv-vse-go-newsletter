package newsletter_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/stretchr/testify/assert"
	// Mocks are defined in create_test.go
)

func TestUpdateHandler(t *testing.T) {
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	mockEditorRepo := new(MockEditorRepository)

	httpHandler := h.UpdateHandler(mockService, mockEditorRepo)

	newsletterID := "nl-xyz-789"
	editorID := "editor-abc-123"

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: editorID}, nil).Once()

		updateName := "Updated Name"
		updateDesc := "Updated Description"
		payload := h.UpdateNewsletterRequest{
			Name:        &updateName,
			Description: &updateDesc,
		}
		expectedNewsletter := &repository.Newsletter{
			ID:          newsletterID,
			EditorID:    editorID,
			Name:        updateName,
			Description: updateDesc,
			UpdatedAt:   time.Now(),
		}
		mockService.On("UpdateNewsletter", newsletterID, editorID, &updateName, &updateDesc).Return(expectedNewsletter, nil).Once()

		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/"+newsletterID, bytes.NewReader(bodyBytes))
		req.SetPathValue("id", newsletterID) // For Go 1.22 PathValue
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resultResp repository.Newsletter
		json.Unmarshal(rr.Body.Bytes(), &resultResp)
		assert.Equal(t, expectedNewsletter.Name, resultResp.Name)

		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/"+newsletterID, nil)
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - Missing Path ID (simulated by handler check)", func(t *testing.T) {
		// Test handler's internal check for newsletterID == ""
		// This is if r.PathValue("id") somehow returns empty, though usually router prevents this.
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "", nil } // Mock to pass auth for this specific check
		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/", nil) // No ID in path for PathValue
		req.SetPathValue("id", "") // Explicitly set empty path value
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		// No mocks expected to be called for service/repo here
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt error")
		}
		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/"+newsletterID, bytes.NewReader([]byte(`{}`))) 
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - Editor Not Found", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(nil, errors.New("editor not found")).Once()

		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/"+newsletterID, bytes.NewReader([]byte(`{}`))) 
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid JSON Body", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "test-uid", nil }
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-uid").Return(&repository.Editor{ID: editorID}, nil).Once()

		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/"+newsletterID, bytes.NewReader([]byte(`{"name": "Test", desc`)))
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Empty Name if Provided", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "test-uid", nil }
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-uid").Return(&repository.Editor{ID: editorID}, nil).Once()

		emptyName := ""
		payload := h.UpdateNewsletterRequest{Name: &emptyName}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/"+newsletterID, bytes.NewReader(bodyBytes))
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Service sql.ErrNoRows (Not Found/Forbidden)", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "test-uid", nil }
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-uid").Return(&repository.Editor{ID: editorID}, nil).Once()

		updateName := "Valid Name"
		payload := h.UpdateNewsletterRequest{Name: &updateName}
		mockService.On("UpdateNewsletter", newsletterID, editorID, &updateName, (*string)(nil)).Return(nil, sql.ErrNoRows).Once()

		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/"+newsletterID, bytes.NewReader(bodyBytes))
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Service ErrNewsletterNameTaken", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "test-uid", nil }
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-uid").Return(&repository.Editor{ID: editorID}, nil).Once()

		conflictName := "Taken Name"
		payload := h.UpdateNewsletterRequest{Name: &conflictName}
		mockService.On("UpdateNewsletter", newsletterID, editorID, &conflictName, (*string)(nil)).Return(nil, service.ErrNewsletterNameTaken).Once()

		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/api/newsletters/"+newsletterID, bytes.NewReader(bodyBytes))
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusConflict, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})
} 