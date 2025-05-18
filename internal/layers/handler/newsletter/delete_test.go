package newsletter_test

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/stretchr/testify/assert"
	// Mocks are defined in create_test.go
)

func TestDeleteHandler(t *testing.T) {
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)
	mockEditorRepo := new(MockEditorRepository)

	httpHandler := h.DeleteHandler(mockService, mockEditorRepo)

	newsletterID := "nl-to-delete-456"
	editorID := "editor-def-456"

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: editorID}, nil).Once()
		mockService.On("DeleteNewsletter", newsletterID, editorID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/newsletters/"+newsletterID, nil)
		req.SetPathValue("id", newsletterID) // For Go 1.22 PathValue
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/newsletters/"+newsletterID, nil)
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - Missing Path ID (simulated by handler check)", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "", nil } 
		req := httptest.NewRequest(http.MethodDelete, "/api/newsletters/", nil)
		req.SetPathValue("id", "")
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt error")
		}
		req := httptest.NewRequest(http.MethodDelete, "/api/newsletters/"+newsletterID, nil)
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

		req := httptest.NewRequest(http.MethodDelete, "/api/newsletters/"+newsletterID, nil)
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Service sql.ErrNoRows (Not Found/Forbidden)", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "test-uid", nil }
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-uid").Return(&repository.Editor{ID: editorID}, nil).Once()
		mockService.On("DeleteNewsletter", newsletterID, editorID).Return(sql.ErrNoRows).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/newsletters/"+newsletterID, nil)
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { return "test-uid", nil }
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-uid").Return(&repository.Editor{ID: editorID}, nil).Once()
		mockService.On("DeleteNewsletter", newsletterID, editorID).Return(errors.New("some other error")).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/newsletters/"+newsletterID, nil)
		req.SetPathValue("id", newsletterID)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})
} 