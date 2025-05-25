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

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("DATABASE_URL environment variable is required for tests")
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
	queries := []string{
		"DELETE FROM posts WHERE title LIKE 'TEST_%'",
		"DELETE FROM newsletters WHERE name LIKE 'TEST_%'",
		"DELETE FROM password_reset_tokens WHERE editor_id IN (SELECT id FROM editors WHERE email LIKE 'test_%@example.com')",
		"DELETE FROM editors WHERE email LIKE 'test_%@example.com'",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Logf("Warning: Failed to cleanup with query %s: %v", query, err)
		}
	}
} 