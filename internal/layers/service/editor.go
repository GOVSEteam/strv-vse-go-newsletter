package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

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
	// Create user in Firebase Auth
	userToCreate := &auth.UserToCreate{}
	userToCreate.Email(email)
	userToCreate.Password(password)

	firebaseUser, err := s.authClient.CreateUser(ctx, userToCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firebase user: %w", err)
	}

	// Create editor record in database
	editor, err := s.repo.InsertEditor(ctx, firebaseUser.UID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to create editor in database: %w", err)
	}

	return editor, nil
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