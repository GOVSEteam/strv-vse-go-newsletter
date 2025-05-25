package testutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

type TestConfig struct {
	DatabaseURL    string
	FirebaseConfig string
	TestMode       bool
}

func LoadTestConfig(t *testing.T) *TestConfig {
	// Load .env file before accessing environment variables
	loadEnvFile(t)
	
	// For local testing, prefer DATABASE_PUBLIC_URL over DATABASE_URL
	// since DATABASE_URL often contains internal hostnames not accessible locally
	databaseURL := os.Getenv("DATABASE_PUBLIC_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	
	return &TestConfig{
		DatabaseURL:    databaseURL,
		FirebaseConfig: os.Getenv("FIREBASE_CONFIG"),
		TestMode:       true,
	}
}

func loadEnvFile(t *testing.T) {
	// Get the directory of the current test file
	_, currentFile, _, ok := runtime.Caller(2) // Go up 2 levels to get the actual test file
	if !ok {
		t.Log("Warning: Failed to get current test file path, .env file may not be loaded")
		return
	}
	
	basepath := filepath.Dir(currentFile)
	// Navigate to project root - repository tests are in internal/layers/repository/
	envPath := filepath.Join(basepath, "../../../.env") // For repository tests: internal/layers/repository -> project root
	
	// Try alternative paths if the default doesn't work
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = filepath.Join(basepath, "../../../../.env") // For deeper nested tests
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			envPath = filepath.Join(basepath, "../../.env") // For shallower tests
		}
	}
	
	// Attempt to load .env file
	err := godotenv.Load(envPath)
	if err != nil {
		t.Logf("Note: .env file not loaded from %s (this might be fine in CI): %v", envPath, err)
	} else {
		t.Logf("Successfully loaded .env file from: %s", envPath)
	}
}

func (c *TestConfig) Validate(t *testing.T) {
	if c.DatabaseURL == "" {
		t.Skip("DATABASE_URL is not set, skipping database-dependent tests")
		return
	}
	
	// Check if DATABASE_URL contains placeholder values
	if strings.Contains(c.DatabaseURL, "YOUR_PASSWORD_HERE") || 
	   strings.Contains(c.DatabaseURL, "YOUR_INTERNAL_HOST_HERE") ||
	   strings.Contains(c.DatabaseURL, "YOUR_HOST_HERE") {
		t.Skip("DATABASE_URL contains placeholder values, skipping database-dependent tests. Please update .env with real database credentials.")
		return
	}
}

func (c *TestConfig) ValidateWithFirebase(t *testing.T) {
	c.Validate(t) // First validate database
	
	if c.FirebaseConfig == "" {
		t.Skip("FIREBASE_CONFIG is not set, skipping Firebase-dependent tests")
		return
	}
} 