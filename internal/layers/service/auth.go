package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
)

const (
	firebasePasswordResetEndpoint = "https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode?key="
	defaultRequestTimeout         = 10 * time.Second
)

// PasswordResetService defines an interface for sending password reset emails.
type PasswordResetService interface {
	SendPasswordResetEmail(ctx context.Context, email string) error
}

// --- FirebasePasswordResetService ---

// FirebasePasswordResetServiceConfig holds configuration for the FirebasePasswordResetService.
type FirebasePasswordResetServiceConfig struct {
	APIKey     string // Firebase Web API Key
	HTTPClient *http.Client
	Logger     *log.Logger
}

// FirebasePasswordResetService implements PasswordResetService using the Firebase Auth REST API.
type FirebasePasswordResetService struct {
	apiKey     string
	httpClient *http.Client
	logger     *log.Logger
}

// NewFirebasePasswordResetService creates a new FirebasePasswordResetService.
func NewFirebasePasswordResetService(config FirebasePasswordResetServiceConfig) (*FirebasePasswordResetService, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("firebase api key is required")
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: defaultRequestTimeout}
	}

	logger := config.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "[FirebasePasswordResetService] ", log.LstdFlags)
	}

	return &FirebasePasswordResetService{
		apiKey:     config.APIKey,
		httpClient: client,
		logger:     logger,
	}, nil
}

type firebaseOobRequest struct {
	RequestType string `json:"requestType"`
	Email       string `json:"email"`
}

type firebaseOobResponse struct {
	Email string `json:"email"` // Email of the user to whom the password reset email was sent.
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
			Domain  string `json:"domain"`
			Reason  string `json:"reason"`
		} `json:"errors"`
	} `json:"error,omitempty"`
}

// SendPasswordResetEmail sends a password reset email via Firebase Auth REST API.
func (s *FirebasePasswordResetService) SendPasswordResetEmail(ctx context.Context, email string) error {
	// 1. Validate email format
	if _, err := mail.ParseAddress(email); err != nil {
		s.logger.Printf("Invalid email format for password reset: %s, error: %v", email, err)
		return fmt.Errorf("%w: %v", apperrors.ErrInvalidEmail, err)
	}

	// 2. Prepare request to Firebase
	requestPayload := firebaseOobRequest{
		RequestType: "PASSWORD_RESET",
		Email:       email,
	}
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		s.logger.Printf("Failed to marshal Firebase password reset request for email %s: %v", email, err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := firebasePasswordResetEndpoint + s.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.logger.Printf("Failed to create Firebase password reset request for email %s: %v", email, err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	s.logger.Printf("Sending password reset request to Firebase for email: %s", email)

	// 3. Make HTTP call
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Printf("Firebase password reset HTTP request failed for email %s: %v", email, err)
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 4. Handle response
	var firebaseResp firebaseOobResponse
	if err := json.NewDecoder(resp.Body).Decode(&firebaseResp); err != nil {
		s.logger.Printf("Failed to decode Firebase password reset response for email %s (status: %d): %v", email, resp.StatusCode, err)
		return fmt.Errorf("failed to decode firebase response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Printf("Firebase password reset request for email %s failed with status %d, response: %+v", email, resp.StatusCode, firebaseResp)
		if firebaseResp.Error != nil {
			switch firebaseResp.Error.Message {
			case "EMAIL_NOT_FOUND":
				return apperrors.ErrEditorNotFound // Or a more generic apperrors.ErrNotFound
			case "INVALID_EMAIL":
				return apperrors.ErrInvalidEmail
			case "MISSING_EMAIL":
				return fmt.Errorf("%w: email is missing in request to Firebase", apperrors.ErrBadRequest)
			// Add more specific Firebase error mappings as needed
			default:
				return fmt.Errorf("firebase error (%s): %w", firebaseResp.Error.Message, apperrors.ErrInternal)
			}
		}
		return fmt.Errorf("firebase request failed with status %d: %w", resp.StatusCode, apperrors.ErrInternal)
	}

	s.logger.Printf("Successfully sent password reset email request to Firebase for: %s", email)
	return nil
}

// --- MockPasswordResetService ---

// MockPasswordResetService is a mock implementation for testing.
type MockPasswordResetService struct {
	SentToEmail string
	ShouldError error
	logger      *log.Logger
}

// NewMockPasswordResetService creates a new MockPasswordResetService.
func NewMockPasswordResetService(logger *log.Logger) *MockPasswordResetService {
	if logger == nil {
		logger = log.New(os.Stderr, "[MockPasswordResetService] ", log.LstdFlags)
	}
	return &MockPasswordResetService{logger: logger}
}

// SendPasswordResetEmail mocks sending an email, stores the email, and returns a configured error.
func (m *MockPasswordResetService) SendPasswordResetEmail(ctx context.Context, email string) error {
	select {
	case <-ctx.Done():
		m.logger.Printf("Context cancelled before sending mock password reset to %s: %v", email, ctx.Err())
		return ctx.Err()
	default:
	}

	// Validate email format (as the real service would)
	if _, err := mail.ParseAddress(email); err != nil {
		m.logger.Printf("Mock: Invalid email format for password reset: %s, error: %v", email, err)
		return fmt.Errorf("%w: %v", apperrors.ErrInvalidEmail, err)
	}

	m.logger.Printf("Mock: Attempting to send password reset email to: %s", email)
	m.SentToEmail = email
	if m.ShouldError != nil {
		m.logger.Printf("Mock: Returning pre-configured error: %v", m.ShouldError)
		return m.ShouldError
	}
	m.logger.Printf("Mock: Successfully 'sent' password reset email to: %s", email)
	return nil
} 