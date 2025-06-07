package testutils

import (
	"context"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	// Load configuration using centralized config system
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}

	// Use the same database URL resolution logic as the main application
	dbURL := cfg.GetDatabaseURL()
	if dbURL == "" {
		t.Fatal("DATABASE_URL is required for tests but not configured")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return pool
}

func CleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	// Clean up in reverse dependency order to avoid foreign key constraint violations
	queries := []string{
		// First, clean up posts (depends on newsletters)
		"DELETE FROM posts WHERE title LIKE 'TEST_%'",
		// Then clean up newsletters (depends on editors)  
		"DELETE FROM newsletters WHERE name LIKE 'TEST_%'",
		// Finally, clean up editors
		"DELETE FROM editors WHERE email LIKE 'test_%@example.com'",
		// Also clean up any editors with TEST_ firebase UIDs
		"DELETE FROM editors WHERE firebase_uid LIKE 'TEST_%'",
	}

	for _, query := range queries {
		if _, err := pool.Exec(context.Background(), query); err != nil {
			t.Logf("Warning: Failed to cleanup with query %s: %v", query, err)
		}
	}
} 