package testutils

import (
	"database/sql"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func SetupTestDB(t *testing.T) *sql.DB {
	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		t.Logf("Warning: .env file not found")
	}

	// For local testing, prefer DATABASE_PUBLIC_URL over DATABASE_URL
	// since DATABASE_URL often contains internal hostnames not accessible locally
	dbURL := os.Getenv("DATABASE_PUBLIC_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	
	if dbURL == "" {
		t.Fatal("DATABASE_URL or DATABASE_PUBLIC_URL environment variable is required for tests")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

func CleanupTestData(t *testing.T, db *sql.DB) {
	// Clean up in reverse dependency order to avoid foreign key constraint violations
	queries := []string{
		// First, clean up posts (depends on newsletters)
		"DELETE FROM posts WHERE title LIKE 'TEST_%'",
		// Then clean up newsletters (depends on editors)  
		"DELETE FROM newsletters WHERE name LIKE 'TEST_%'",
		// Clean up password reset tokens (depends on editors)
		"DELETE FROM password_reset_tokens WHERE editor_id IN (SELECT id FROM editors WHERE email LIKE 'test_%@example.com')",
		// Finally, clean up editors
		"DELETE FROM editors WHERE email LIKE 'test_%@example.com'",
		// Also clean up any editors with TEST_ firebase UIDs
		"DELETE FROM editors WHERE firebase_uid LIKE 'TEST_%'",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Logf("Warning: Failed to cleanup with query %s: %v", query, err)
		}
	}
} 