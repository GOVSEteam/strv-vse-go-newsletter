package service

import (
	"bytes"
	"context"
	"encoding/json"
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/firebase-auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"net/http"
	"os"
)

type SignInResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	Email        string `json:"email"`
	LocalID      string `json:"localId"`
}

type EditorService interface {
	SignUp(email, password string) (*repository.Editor, error)
	SignIn(email, password string) (*SignInResponse, error)
}

type editorService struct {
	repo repository.EditorRepository
}

func NewEditorService(repo repository.EditorRepository) EditorService {
	return &editorService{repo: repo}
}

func (s *editorService) SignUp(email, password string) (*repository.Editor, error) {
	client := firebase_auth.GetAuthClient()
	params := (&auth.UserToCreate{}).Email(email).Password(password)
	user, err := client.CreateUser(context.Background(), params)
	if err != nil {
		return nil, err
	}
	return s.repo.InsertEditor(user.UID, email)
}

func (s *editorService) SignIn(email, password string) (*SignInResponse, error) {
	apiKey := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS env var not set (should be Firebase Web API Key)")
	}
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", apiKey)
	payload := map[string]interface{}{
		"email":             email,
		"password":          password,
		"returnSecureToken": true,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("firebase sign-in failed: %v", errResp)
	}
	var out SignInResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
