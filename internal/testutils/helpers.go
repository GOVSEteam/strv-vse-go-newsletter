package testutils

import (
	"fmt"
	"strings"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
)

// Test data generators
func CreateTestEditor(suffix int) *repository.Editor {
	timestamp := time.Now().Unix()
	return &repository.Editor{
		Email:       fmt.Sprintf("test_editor_%d_%d@example.com", suffix, timestamp),
		FirebaseUID: fmt.Sprintf("TEST_firebase_uid_%d_%d", suffix, timestamp),
	}
}

func CreateTestNewsletter(editorID string, suffix int) *repository.Newsletter {
	return &repository.Newsletter{
		EditorID:    editorID,
		Name:        fmt.Sprintf("TEST_Newsletter_%03d", suffix),
		Description: fmt.Sprintf("TEST_Description for testing newsletter %d", suffix),
	}
}

func CreateTestPost(newsletterID uuid.UUID, suffix int) *models.Post {
	return &models.Post{
		NewsletterID: newsletterID,
		Title:        fmt.Sprintf("TEST_Post_%03d", suffix),
		Content:      fmt.Sprintf("TEST_Content for testing post %d", suffix),
	}
}

func CreateTestSubscriber(newsletterID string, suffix int) *models.Subscriber {
	timestamp := time.Now().Unix()
	now := time.Now()
	return &models.Subscriber{
		Email:        fmt.Sprintf("test_subscriber_%d_%d@example.com", suffix, timestamp),
		NewsletterID: newsletterID,
		ConfirmedAt:  &now,
	}
}

// Test identification helpers
func IsTestData(email, name, title string) bool {
	return strings.HasPrefix(email, "test_") ||
		strings.HasPrefix(name, "TEST_") ||
		strings.HasPrefix(title, "TEST_")
} 