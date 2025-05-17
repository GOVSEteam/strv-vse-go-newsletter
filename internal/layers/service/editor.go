package service

import (
	"context"
	"firebase.google.com/go/v4/auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/firebase-auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
)

type EditorService interface {
	SignUp(email, password string) (*repository.Editor, error)
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
