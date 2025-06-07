package testutils

import (
	"strings"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/config"
)

// LoadTestConfig loads and validates configuration for tests.
// Uses the same config system as production for consistency.
func LoadTestConfig(t *testing.T) *config.Config {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}
	
	ValidateTestConfig(t, cfg)
	return cfg
}

// ValidateTestConfig ensures test configuration is properly set up.
func ValidateTestConfig(t *testing.T, cfg *config.Config) {
	dbURL := cfg.GetDatabaseURL()
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set, skipping database-dependent tests")
		return
	}
	
	// Check if DATABASE_URL contains placeholder values
	if strings.Contains(dbURL, "YOUR_PASSWORD_HERE") || 
	   strings.Contains(dbURL, "YOUR_INTERNAL_HOST_HERE") ||
	   strings.Contains(dbURL, "YOUR_HOST_HERE") {
		t.Skip("DATABASE_URL contains placeholder values, skipping database-dependent tests. Please update .env with real database credentials.")
		return
	}
}

// ValidateTestConfigWithFirebase validates both database and Firebase configuration.
func ValidateTestConfigWithFirebase(t *testing.T, cfg *config.Config) {
	ValidateTestConfig(t, cfg) // First validate database
	
	if cfg.FirebaseServiceAccount == "" {
		t.Skip("FIREBASE_SERVICE_ACCOUNT is not set, skipping Firebase-dependent tests")
		return
	}
} 