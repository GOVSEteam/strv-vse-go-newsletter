package service

import (
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
)

type NewsletterServiceInterface interface {
	ListNewsletters() ([]repository.Newsletter, error)
	CreateNewsletter(editorID, name, description string) (*repository.Newsletter, error)
}

type newsletterService struct {
	repo repository.NewsletterRepository
}

func NewsletterService(repo repository.NewsletterRepository) NewsletterServiceInterface {
	return &newsletterService{repo: repo}
}

func (s *newsletterService) ListNewsletters() ([]repository.Newsletter, error) {
	return s.repo.ListNewsletters()
}

func (s *newsletterService) CreateNewsletter(editorID, name, description string) (*repository.Newsletter, error) {
	return s.repo.CreateNewsletter(editorID, name, description)
}
