package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// Test the most critical part: data mapping from database to domain model
func TestDbNewsletter_ToModel(t *testing.T) {
	tests := []struct {
		name         string
		dbNewsletter dbNewsletter
		expected     models.Newsletter
	}{
		{
			name: "complete newsletter mapping",
			dbNewsletter: dbNewsletter{
				ID:          "newsletter_123",
				EditorID:    "editor_456",
				Name:        "Tech Weekly",
				Description: "A weekly tech newsletter",
				CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: models.Newsletter{
				ID:          "newsletter_123",
				EditorID:    "editor_456",
				Name:        "Tech Weekly",
				Description: "A weekly tech newsletter",
				CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "newsletter with empty description",
			dbNewsletter: dbNewsletter{
				ID:          "newsletter_456",
				EditorID:    "editor_789",
				Name:        "Simple Newsletter",
				Description: "",
				CreatedAt:   time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: models.Newsletter{
				ID:          "newsletter_456",
				EditorID:    "editor_789",
				Name:        "Simple Newsletter",
				Description: "",
				CreatedAt:   time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 2, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "newsletter with unicode characters",
			dbNewsletter: dbNewsletter{
				ID:          "newsletter_unicode",
				EditorID:    "editor_unicode",
				Name:        "📧 Newsletter 🚀",
				Description: "Newsletter with émojis and speçial chars",
				CreatedAt:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: models.Newsletter{
				ID:          "newsletter_unicode",
				EditorID:    "editor_unicode",
				Name:        "📧 Newsletter 🚀",
				Description: "Newsletter with émojis and speçial chars",
				CreatedAt:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dbNewsletter.toModel()

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.EditorID, result.EditorID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Description, result.Description)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
		})
	}
}

// Test repository constructor
func TestNewPostgresNewsletterRepo(t *testing.T) {
	t.Run("creates newsletter repository", func(t *testing.T) {
		// We can't easily test with a real DB pool without setup
		// but we can test the constructor returns the right type
		repo := NewPostgresNewsletterRepo(nil)
		assert.NotNil(t, repo)
		
		// Verify it implements the interface
		var _ NewsletterRepository = repo
	})
}

// Test critical database struct mapping edge cases
func TestDbNewsletter_EdgeCases(t *testing.T) {
	t.Run("handles zero time values", func(t *testing.T) {
		dbNewsletter := dbNewsletter{
			ID:          "test_id",
			EditorID:    "test_editor",
			Name:        "Test Newsletter",
			Description: "Test Description",
			CreatedAt:   time.Time{}, // Zero time
			UpdatedAt:   time.Time{}, // Zero time
		}

		result := dbNewsletter.toModel()

		assert.Equal(t, "test_id", result.ID)
		assert.Equal(t, "test_editor", result.EditorID)
		assert.True(t, result.CreatedAt.IsZero())
		assert.True(t, result.UpdatedAt.IsZero())
	})

	t.Run("preserves exact field values", func(t *testing.T) {
		// Test that no data transformation occurs during mapping
		dbNewsletter := dbNewsletter{
			ID:          "exact_id_123",
			EditorID:    "exact_editor_456",
			Name:        "  Exact Name  ", // With whitespace
			Description: "Line 1\nLine 2\nLine 3", // With newlines
			CreatedAt:   time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 987654321, time.UTC),
		}

		result := dbNewsletter.toModel()

		// Verify no data transformation
		assert.Equal(t, "exact_id_123", result.ID)
		assert.Equal(t, "exact_editor_456", result.EditorID)
		assert.Equal(t, "  Exact Name  ", result.Name) // Whitespace preserved
		assert.Equal(t, "Line 1\nLine 2\nLine 3", result.Description) // Newlines preserved
		assert.Equal(t, 123456789, result.CreatedAt.Nanosecond())
		assert.Equal(t, 987654321, result.UpdatedAt.Nanosecond())
	})

	t.Run("handles maximum length strings", func(t *testing.T) {
		nameBytes := make([]byte, 100)
		for i := range nameBytes {
			nameBytes[i] = 'A'
		}
		longName := string(nameBytes)
		
		descBytes := make([]byte, 500)
		for i := range descBytes {
			descBytes[i] = 'B'
		}
		longDescription := string(descBytes)

		dbNewsletter := dbNewsletter{
			ID:          "max_length_test",
			EditorID:    "editor_max",
			Name:        longName,
			Description: longDescription,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		result := dbNewsletter.toModel()

		assert.Equal(t, longName, result.Name)
		assert.Equal(t, longDescription, result.Description)
		assert.Len(t, result.Name, 100)
		assert.Len(t, result.Description, 500)
	})
}

// Test data integrity scenarios that could occur in real database operations
func TestNewsletterRepository_DataIntegrity(t *testing.T) {
	t.Run("newsletter ID consistency", func(t *testing.T) {
		// Test that the same data in -> same data out
		originalData := dbNewsletter{
			ID:          "consistency_test_123",
			EditorID:    "editor_consistency",
			Name:        "Consistency Test Newsletter",
			Description: "Testing data consistency",
			CreatedAt:   time.Date(2023, 6, 15, 9, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC),
		}

		// Convert to model and verify
		model1 := originalData.toModel()
		
		// Create another dbNewsletter with same data
		sameData := dbNewsletter{
			ID:          "consistency_test_123",
			EditorID:    "editor_consistency", 
			Name:        "Consistency Test Newsletter",
			Description: "Testing data consistency",
			CreatedAt:   time.Date(2023, 6, 15, 9, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC),
		}

		model2 := sameData.toModel()

		// Should be identical
		assert.Equal(t, model1, model2)
	})

	t.Run("field independence", func(t *testing.T) {
		// Test that changing one field doesn't affect others
		base := dbNewsletter{
			ID:          "field_test",
			EditorID:    "editor_field", 
			Name:        "Original Name",
			Description: "Original Description",
			CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		modified := base
		modified.Name = "Modified Name"

		baseModel := base.toModel()
		modifiedModel := modified.toModel()

		// Only name should be different
		assert.NotEqual(t, baseModel.Name, modifiedModel.Name)
		assert.Equal(t, baseModel.ID, modifiedModel.ID)
		assert.Equal(t, baseModel.EditorID, modifiedModel.EditorID)
		assert.Equal(t, baseModel.Description, modifiedModel.Description)
		assert.Equal(t, baseModel.CreatedAt, modifiedModel.CreatedAt)
		assert.Equal(t, baseModel.UpdatedAt, modifiedModel.UpdatedAt)
	})
} 