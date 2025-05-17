package service

import (
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
)

type NewsletterServiceInterface interface {
	ListNewsletters() ([]string, error)
	CreateNewsletter(name string) error
}

type newsletterService struct {
	repo repository.NewsletterRepository
}

func NewsletterService(repo repository.NewsletterRepository) NewsletterServiceInterface {
	return &newsletterService{repo: repo}
}

func (s *newsletterService) ListNewsletters() ([]string, error) {
	return s.repo.ListNewsletters()
}

func (s *newsletterService) CreateNewsletter(name string) error {
	return s.repo.CreateNewsletter(name)
}
