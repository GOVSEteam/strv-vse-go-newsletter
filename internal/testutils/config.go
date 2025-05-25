package testutils

import (
	"os"
	"testing"
)

type TestConfig struct {
	DatabaseURL    string
	FirebaseConfig string
	TestMode       bool
}

func LoadTestConfig(t *testing.T) *TestConfig {
	return &TestConfig{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		FirebaseConfig: os.Getenv("FIREBASE_CONFIG"),
		TestMode:       true,
	}
}

func (c *TestConfig) Validate(t *testing.T) {
	if c.DatabaseURL == "" {
		t.Fatal("DATABASE_URL is required for tests")
	}

	if c.FirebaseConfig == "" {
		t.Fatal("FIREBASE_CONFIG is required for tests")
	}
} 