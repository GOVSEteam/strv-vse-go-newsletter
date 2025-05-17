package service

import (
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
)

type NewsletterService interface {
	ListNewsletters() ([]string, error)
	CreateNewsletter(name string) error
}

type newsletterService struct {
	repo repository.NewsletterRepository
}

func NewNewsletterService(repo repository.NewsletterRepository) NewsletterService {
	return &newsletterService{repo: repo}
}

func (s *newsletterService) ListNewsletters() ([]string, error) {
	return s.repo.ListNewsletters()
}

func (s *newsletterService) CreateNewsletter(name string) error {
	return s.repo.CreateNewsletter(name)
}
