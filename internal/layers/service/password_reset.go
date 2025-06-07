package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// PasswordResetService defines the interface for password reset operations
type PasswordResetService interface {
	SendPasswordResetEmail(ctx context.Context, email string) error
}

// FirebasePasswordResetServiceConfig holds configuration for Firebase password reset service
type FirebasePasswordResetServiceConfig struct {
	APIKey     string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// FirebasePasswordResetService implements PasswordResetService using Firebase Auth API
type FirebasePasswordResetService struct {
	config FirebasePasswordResetServiceConfig
}

// NewFirebasePasswordResetService creates a new Firebase password reset service
func NewFirebasePasswordResetService(config FirebasePasswordResetServiceConfig) (*FirebasePasswordResetService, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("firebase API key is required")
	}
	if config.HTTPClient == nil {
		return nil, fmt.Errorf("HTTP client is required")
	}
	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &FirebasePasswordResetService{
		config: config,
	}, nil
}

// SendPasswordResetEmail sends a password reset email using Firebase Auth API
func (s *FirebasePasswordResetService) SendPasswordResetEmail(ctx context.Context, email string) error {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode?key=%s", s.config.APIKey)
	
	payload := map[string]interface{}{
		"requestType": "PASSWORD_RESET",
		"email":       email,
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send password reset request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		s.config.Logger.Printf("Firebase password reset request failed with status: %d", resp.StatusCode)
		return fmt.Errorf("password reset request failed with status: %d", resp.StatusCode)
	}
	
	s.config.Logger.Printf("Password reset email sent successfully to %s", email)
	return nil
} 