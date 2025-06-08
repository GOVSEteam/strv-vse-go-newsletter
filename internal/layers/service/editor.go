package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// CreateUserRequest represents the data needed to create a new user
type CreateUserRequest struct {
	Email    string
	Password string
}

// CreatedUser represents a newly created user
type CreatedUser struct {
	UID   string
	Email string
}

// AuthClient defines the interface for authentication operations (provider-agnostic).
// This interface abstracts authentication providers to enable testing and follow
// the dependency inversion principle.
type AuthClient interface {
	// CreateUser creates a new user with the given parameters
	CreateUser(ctx context.Context, req CreateUserRequest) (*CreatedUser, error)
}

// FirebaseAuthClient is deprecated - use AuthClient instead
// Keeping for backward compatibility during transition
type FirebaseAuthClient = AuthClient

// SignInResponse represents the response from Firebase sign-in
type SignInResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	Email        string `json:"email"`
	LocalID      string `json:"localId"`
}

// EditorServiceInterface defines the interface for editor operations
type EditorServiceInterface interface {
	SignUp(ctx context.Context, email, password string) (*models.Editor, error)
	SignIn(ctx context.Context, email, password string) (*SignInResponse, error)
}

// editorService implements EditorServiceInterface
type editorService struct {
	repo               repository.EditorRepository
	authClient         FirebaseAuthClient
	httpClient         *http.Client
	firebaseAPIKey     string
	firebaseSignInURL  string
}

// NewEditorService creates a new editor service
func NewEditorService(
	repo repository.EditorRepository,
	authClient FirebaseAuthClient,
	httpClient *http.Client,
	firebaseAPIKey string,
) EditorServiceInterface {
	return &editorService{
		repo:               repo,
		authClient:         authClient,
		httpClient:         httpClient,
		firebaseAPIKey:     firebaseAPIKey,
		firebaseSignInURL:  fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", firebaseAPIKey),
	}
}

// SignUp creates a new editor account
func (s *editorService) SignUp(ctx context.Context, email, password string) (*models.Editor, error) {
	// Create user request
	createReq := CreateUserRequest{
		Email:    email,
		Password: password,
	}

	createdUser, err := s.authClient.CreateUser(ctx, createReq)
	if err != nil {
		// Map auth provider errors to appropriate application errors using centralized error system
		if isAuthConflictError(err) {
			return nil, apperrors.ErrConflict
		}
		if isAuthValidationError(err) {
			return nil, apperrors.WrapValidation(err, "invalid signup data")
		}
		return nil, fmt.Errorf("editor service: failed to create user: %w", err)
	}

	// Create editor record in database
	editor, err := s.repo.InsertEditor(ctx, createdUser.UID, email)
	if err != nil {
		// The repository already returns proper error types
		return nil, fmt.Errorf("editor service: failed to create editor in database: %w", err)
	}

	return editor, nil
}

// isAuthConflictError checks if the auth provider error indicates a user already exists
func isAuthConflictError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "email already exists") ||
		strings.Contains(errStr, "user with the provided email already exists") ||
		strings.Contains(errStr, "email_already_exists") ||
		strings.Contains(errStr, "duplicate")
}

// isAuthValidationError checks if the auth provider error indicates invalid input
func isAuthValidationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "invalid email") ||
		strings.Contains(errStr, "weak password") ||
		strings.Contains(errStr, "password should be at least") ||
		strings.Contains(errStr, "invalid_email") ||
		strings.Contains(errStr, "weak_password")
}

// SignIn authenticates an editor and returns Firebase tokens
func (s *editorService) SignIn(ctx context.Context, email, password string) (*SignInResponse, error) {
	payload := map[string]interface{}{
		"email":             email,
		"password":          password,
		"returnSecureToken": true,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sign-in payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.firebaseSignInURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create sign-in request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to sign in: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sign-in failed with status: %d", resp.StatusCode)
	}

	var signInResp SignInResponse
	if err := json.NewDecoder(resp.Body).Decode(&signInResp); err != nil {
		return nil, fmt.Errorf("failed to decode sign-in response: %w", err)
	}

	return &signInResp, nil
} 