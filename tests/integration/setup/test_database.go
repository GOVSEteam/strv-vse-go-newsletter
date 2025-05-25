package setup

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// SetupTestDatabase creates and configures a test database connection
func SetupTestDatabase(t *testing.T) *sql.DB {
	dbURL := getTestDatabaseURL()
	if dbURL == "" {
		t.Skip("No test database URL configured")
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	// Ensure tables exist (run migrations if needed)
	ensureTablesExist(t, db)

	// Clean any existing test data
	cleanTestData(t, db)

	return db
}

// CleanupTestDatabase cleans up test data and closes the database connection
func CleanupTestDatabase(db *sql.DB) {
	if db != nil {
		// Clean test data before closing
		cleanTestData(nil, db)
	}
}

// getTestDatabaseURL returns the database URL for testing
func getTestDatabaseURL() string {
	// Try test-specific environment variables first
	if testURL := os.Getenv("TEST_DATABASE_URL"); testURL != "" {
		return testURL
	}

	// For Railway and other cloud providers, prioritize public URL for local testing
	if publicURL := os.Getenv("DATABASE_PUBLIC_URL"); publicURL != "" {
		return publicURL
	}

	// Fall back to regular database URL
	dbURL := os.Getenv("DATABASE_URL")
	return dbURL
}

// ensureTablesExist creates necessary tables if they don't exist
func ensureTablesExist(t *testing.T, db *sql.DB) {
	// Check if tables exist, create them if not
	// Note: subscribers are stored in Firestore, not PostgreSQL
	tables := []string{"editors", "newsletters", "posts"}
	
	for _, table := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`
		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil && t != nil {
			t.Logf("Warning: Could not check if table %s exists: %v", table, err)
		}
		
		if !exists && t != nil {
			t.Logf("Warning: Table %s does not exist. Run migrations first.", table)
		}
	}
}

// cleanTestData removes all test data from the database
func cleanTestData(t *testing.T, db *sql.DB) {
	// Clean in reverse dependency order to avoid foreign key constraints
	// Note: subscribers are stored in Firestore and cleaned separately
	cleanQueries := []string{
		"DELETE FROM posts WHERE newsletter_id IN (SELECT id FROM newsletters WHERE editor_id IN (SELECT id FROM editors WHERE email LIKE '%@test.example.com' OR email LIKE '%@integration.test'))",
		"DELETE FROM newsletters WHERE editor_id IN (SELECT id FROM editors WHERE email LIKE '%@test.example.com' OR email LIKE '%@integration.test')",
		"DELETE FROM editors WHERE email LIKE '%@test.example.com' OR email LIKE '%@integration.test'",
	}

	for _, query := range cleanQueries {
		_, err := db.Exec(query)
		if err != nil && t != nil {
			t.Logf("Warning: Failed to clean test data with query %s: %v", query, err)
		}
	}
}

// CreateTestTransaction creates a transaction for test isolation
func CreateTestTransaction(t *testing.T, db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	require.NoError(t, err, "Failed to begin test transaction")
	return tx
}

// RollbackTestTransaction rolls back a test transaction
func RollbackTestTransaction(tx *sql.Tx) {
	if tx != nil {
		tx.Rollback()
	}
}

// TestDatabaseConfig holds configuration for test database
type TestDatabaseConfig struct {
	UseTransactions bool // Whether to use transactions for test isolation
	CleanupAfter    bool // Whether to cleanup after each test
	SkipMigrations  bool // Whether to skip migration checks
}

// DefaultTestDatabaseConfig returns default configuration for test database
func DefaultTestDatabaseConfig() TestDatabaseConfig {
	return TestDatabaseConfig{
		UseTransactions: false, // Disabled by default as it can interfere with HTTP handlers
		CleanupAfter:    true,
		SkipMigrations:  false,
	}
} 