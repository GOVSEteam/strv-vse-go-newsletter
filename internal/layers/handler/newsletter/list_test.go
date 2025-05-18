package newsletter_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	// Mocks are already defined in create_test.go in the same package
)

func TestListHandler(t *testing.T) {
	originalVerifyJWT := auth.VerifyFirebaseJWT
	defer func() { auth.VerifyFirebaseJWT = originalVerifyJWT }()

	mockService := new(MockNewsletterService)    // Defined in create_test.go
	mockEditorRepo := new(MockEditorRepository) // Defined in create_test.go

	httpHandler := h.ListHandler(mockService, mockEditorRepo)

	t.Run("Success", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		expectedNewsletters := []repository.Newsletter{
			{ID: "nl-1", EditorID: "editor-123", Name: "N1", CreatedAt: time.Now()},
			{ID: "nl-2", EditorID: "editor-123", Name: "N2", CreatedAt: time.Now().Add(-time.Hour)},
		}
		expectedTotal := 20
		mockService.On("ListNewslettersByEditorID", mock.Anything, "editor-123", h.DefaultLimit, h.DefaultOffset).Return(expectedNewsletters, expectedTotal, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var responseBody h.PaginatedNewslettersResponse
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Equal(t, expectedTotal, responseBody.Total)
		assert.Equal(t, h.DefaultLimit, responseBody.Limit)
		assert.Equal(t, h.DefaultOffset, responseBody.Offset)
		assert.Len(t, responseBody.Data, len(expectedNewsletters))
		assert.Equal(t, expectedNewsletters[0].Name, responseBody.Data[0].Name)

		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Success with limit and offset params", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		customLimit, customOffset := 5, 10
		mockService.On("ListNewslettersByEditorID", mock.Anything, "editor-123", customLimit, customOffset).Return([]repository.Newsletter{}, 0, nil).Once()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/newsletters?limit=%d&offset=%d", customLimit, customOffset), nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody h.PaginatedNewslettersResponse
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Equal(t, customLimit, responseBody.Limit)
		assert.Equal(t, customOffset, responseBody.Offset)

		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Error - JWT Verification Fails", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("jwt error")
		}
		req := httptest.NewRequest(http.MethodGet, "/api/newsletters", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - Editor Not Found", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(nil, errors.New("editor not found")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid Limit Param", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) { // Auth must pass to reach param parsing
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters?limit=abc", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockEditorRepo.AssertExpectations(t)
	})

	t.Run("Error - Service Generic Error", func(t *testing.T) {
		auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "test-firebase-uid", nil
		}
		mockEditorRepo.On("GetEditorByFirebaseUID", "test-firebase-uid").Return(&repository.Editor{ID: "editor-123"}, nil).Once()
		mockService.On("ListNewslettersByEditorID", mock.Anything, "editor-123", h.DefaultLimit, h.DefaultOffset).Return(([]repository.Newsletter)(nil), 0, errors.New("db error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/newsletters", nil)
		rr := httptest.NewRecorder()
		httpHandler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
		mockEditorRepo.AssertExpectations(t)
	})
}
