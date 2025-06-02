package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"firebase.google.com/go/v4/auth"
	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

const (
	MinPasswordLength = 6
)

// EmailRegex defines a simple regex for email validation.
var EmailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

type SignInResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	Email        string `json:"email"`
	LocalID      string `json:"localId"`
	ExpiresIn    string `json:"expiresIn"`       // Added from typical Firebase response
	Registered   bool   `json:"registered"`      // Added from typical Firebase response
}

// FirebaseErrorDetail is part of the Firebase error response structure.
type FirebaseErrorDetail struct {
	Message string `json:"message"`
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
}

// FirebaseErrorResponse is the structure for Firebase REST API errors.
type FirebaseErrorResponse struct {
	Error struct {
		Code    int                   `json:"code"`
		Message string                `json:"message"`
		Errors  []FirebaseErrorDetail `json:"errors"`
	} `json:"error"`
}

type EditorServiceInterface interface {
	SignUp(ctx context.Context, email, password string) (*models.Editor, error)
	SignIn(ctx context.Context, email, password string) (*SignInResponse, error)
	// GetEditorByFirebaseUID might be useful here if needed by other services/handlers directly
	// GetEditorByFirebaseUID(ctx context.Context, firebaseUID string) (*models.Editor, error)
}

type editorService struct {
	repo             repository.EditorRepository
	authClient       *auth.Client
	httpClient       *http.Client
	firebaseAPIKey   string
	firebaseSignInURL string
}

func NewEditorService(
	repo repository.EditorRepository,
	authClient *auth.Client,
	httpClient *http.Client,
	firebaseAPIKey string,
) EditorServiceInterface {
	return &editorService{
		repo:             repo,
		authClient:       authClient,
		httpClient:       httpClient,
		firebaseAPIKey:   firebaseAPIKey,
		firebaseSignInURL: fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", firebaseAPIKey),
	}
}

func (s *editorService) SignUp(ctx context.Context, email, password string) (*models.Editor, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if email == "" {
		return nil, fmt.Errorf("service: SignUp: %w: email cannot be empty", apperrors.ErrValidation)
	}
	// Basic regex check + net/mail for better validation
	if _, err := mail.ParseAddress(email); err != nil || !EmailRegex.MatchString(email) {
		return nil, fmt.Errorf("service: SignUp: %w: %s", apperrors.ErrInvalidEmail, email)
	}

	if password == "" {
		return nil, fmt.Errorf("service: SignUp: %w: password cannot be empty", apperrors.ErrValidation)
	}
	if len(password) < MinPasswordLength {
		return nil, fmt.Errorf("service: SignUp: %w: password must be at least %d characters", apperrors.ErrPasswordTooShort, MinPasswordLength)
	}

	params := (&auth.UserToCreate{}).Email(email).Password(password)
	firebaseUser, err := s.authClient.CreateUser(ctx, params)
	if err != nil {
		if auth.IsEmailAlreadyExists(err) {
			return nil, fmt.Errorf("service: SignUp: CreateUser: %w: email '%s' already exists", apperrors.ErrConflict, email)
		}
		// Other auth errors could be: IsInvalidEmail, IsUnsupportedPassword, IsInternalError etc.
		// For simplicity, map general Firebase errors to internal or a specific auth provider error.
		return nil, fmt.Errorf("service: SignUp: CreateUser: %w: %v", apperrors.ErrInternal, err) // Or a more specific "auth provider error"
	}

	// Insert into our database
	editor, err := s.repo.InsertEditor(ctx, firebaseUser.UID, email)
	if err != nil {
		// This could be a conflict if somehow UID/email is already in DB (shouldn't happen if Firebase part is atomic for email)
		// Or a general DB error.
		// The repository should wrap pgx errors into apperrors.ErrConflict or apperrors.ErrInternal
		return nil, fmt.Errorf("service: SignUp: InsertEditor: %w", err)
	}

	return editor, nil
}

func (s *editorService) SignIn(ctx context.Context, email, password string) (*SignInResponse, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	if email == "" {
		return nil, fmt.Errorf("service: SignIn: %w: email cannot be empty", apperrors.ErrValidation)
	}
	if password == "" {
		return nil, fmt.Errorf("service: SignIn: %w: password cannot be empty", apperrors.ErrValidation)
	}
	// No need to validate email format here as Firebase will do it.
	// No need to validate password length here as Firebase will do it.

	payload := map[string]interface{}{
		"email":             email,
		"password":          password,
		"returnSecureToken": true,
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("service: SignIn: marshal payload: %w: %v", apperrors.ErrInternal, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.firebaseSignInURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("service: SignIn: create request: %w: %v", apperrors.ErrInternal, err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Consider setting a timeout on the httpClient if not already configured globally.
	// For example: s.httpClient.Timeout = 10 * time.Second
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("service: SignIn: http client Do: %w: %v", apperrors.ErrInternal, err) // Could be network error, timeout
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var firebaseErrResp FirebaseErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&firebaseErrResp); err == nil {
			// Try to map specific Firebase error messages
			if firebaseErrResp.Error.Message == "EMAIL_NOT_FOUND" ||
				firebaseErrResp.Error.Message == "INVALID_PASSWORD" || // Firebase v1 uses INVALID_PASSWORD
				firebaseErrResp.Error.Message == "INVALID_LOGIN_CREDENTIALS" { // Newer Firebase versions might use this
				return nil, fmt.Errorf("service: SignIn: %w: invalid email or password", apperrors.ErrUnauthorized)
			}
			if firebaseErrResp.Error.Message == "USER_DISABLED" {
				return nil, fmt.Errorf("service: SignIn: %w: user account disabled", apperrors.ErrForbidden)
			}
			// Fallback for other Firebase errors
			return nil, fmt.Errorf("service: SignIn: auth provider failed: %s (code %d): %w", firebaseErrResp.Error.Message, resp.StatusCode, apperrors.ErrInternal)
		}
		// If decoding Firebase error fails, return a generic error
		return nil, fmt.Errorf("service: SignIn: auth provider request failed with status %d: %w", resp.StatusCode, apperrors.ErrInternal)
	}

	var out SignInResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("service: SignIn: decode response: %w: %v", apperrors.ErrInternal, err)
	}

	return &out, nil
}

// Example of GetEditorByFirebaseUID if needed by other parts (e.g. auth middleware directly using service)
// func (s *editorService) GetEditorByFirebaseUID(ctx context.Context, firebaseUID string) (*models.Editor, error) {
// 	if firebaseUID == "" {
// 		return nil, fmt.Errorf("service: GetEditorByFirebaseUID: %w: firebaseUID cannot be empty", apperrors.ErrValidation)
// 	}
// 	editor, err := s.repo.GetEditorByFirebaseUID(ctx, firebaseUID)
// 	if err != nil {
// 		if errors.Is(err, apperrors.ErrEditorNotFound) { // Assuming repo returns this
// 			return nil, fmt.Errorf("service: GetEditorByFirebaseUID: %w", apperrors.ErrEditorNotFound)
// 		}
// 		return nil, fmt.Errorf("service: GetEditorByFirebaseUID: %w", err) // Other internal errors
// 	}
// 	return editor, nil
// }
