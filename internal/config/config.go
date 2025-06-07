package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application
type Config struct {
	// Database configuration
	DatabaseURL            string
	DatabasePublicURL      string

	// Firebase configuration  
	FirebaseServiceAccount string
	FirebaseAPIKey         string

	// Email configuration
	GoogleAppPassword string
	EmailFrom         string
	SMTPHost          string
	SMTPPort          string

	// Application configuration
	AppBaseURL string
	Port       int

	// Environment
	RailwayEnvironment string
}

// Load reads configuration from environment variables and validates required fields
func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		DatabasePublicURL:      os.Getenv("DATABASE_PUBLIC_URL"),
		FirebaseServiceAccount: getFirebaseServiceAccount(),
		FirebaseAPIKey:         os.Getenv("FIREBASE_API_KEY"),
		GoogleAppPassword:      os.Getenv("GOOGLE_APP_PASSWORD"),
		EmailFrom:              os.Getenv("EMAIL_FROM"),
		AppBaseURL:             os.Getenv("APP_BASE_URL"),
		RailwayEnvironment:     os.Getenv("RAILWAY_ENVIRONMENT"),
	}

	// Set SMTP defaults
	config.SMTPHost = getEnvWithDefault("SMTP_HOST", "smtp.gmail.com")
	config.SMTPPort = getEnvWithDefault("SMTP_PORT", "587")

	// Parse port with default
	port, err := strconv.Atoi(getEnvWithDefault("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	config.Port = port

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks required configuration fields
func (c *Config) validate() error {
	required := map[string]string{
		"DATABASE_URL":             c.DatabaseURL,
		"FIREBASE_SERVICE_ACCOUNT": c.FirebaseServiceAccount,
		"FIREBASE_API_KEY":         c.FirebaseAPIKey,
		"APP_BASE_URL":             c.AppBaseURL,
	}

	for field, value := range required {
		if value == "" {
			return fmt.Errorf("%s is required", field)
		}
	}

	return nil
}

// GetDatabaseURL returns the appropriate database URL
func (c *Config) GetDatabaseURL() string {
	if c.DatabasePublicURL != "" && c.RailwayEnvironment == "" {
		return c.DatabasePublicURL
	}
	return c.DatabaseURL
}

// getEnvWithDefault returns environment variable value or default if empty
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getFirebaseServiceAccount returns the Firebase service account JSON.
// It tries multiple sources in order: file, base64 env var, plain env var.
func getFirebaseServiceAccount() string {
	// Try reading from file first (most reliable)
	if data, err := os.ReadFile("firebase-service-account.json"); err == nil {
		return string(data)
	}
	
	// Try base64 encoded version
	if encoded := os.Getenv("FIREBASE_SERVICE_ACCOUNT_BASE64"); encoded != "" {
		if decoded, err := base64.StdEncoding.DecodeString(encoded); err == nil {
			return string(decoded)
		}
	}
	
	// Handle plain JSON from environment variable
	if plainJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT"); plainJSON != "" {
		// If the JSON contains actual newlines (multiline), we need to handle it properly
		// This commonly happens when the environment variable is set from a multiline file
		if strings.Contains(plainJSON, "\n") {
			// For multiline JSON, we need to ensure proper JSON escaping
			// The most common issue is unescaped newlines in the private_key field
			return fixFirebaseJSONNewlines(plainJSON)
		}
		// If it contains literal \n sequences, replace them with actual newlines
		return strings.ReplaceAll(plainJSON, "\\n", "\n")
	}
	
	return ""
}

// fixFirebaseJSONNewlines fixes JSON that contains unescaped newlines, particularly
// in the private_key field of Firebase service account JSON
func fixFirebaseJSONNewlines(jsonStr string) string {
	// Use a more robust approach: parse the JSON structure and properly escape the private_key
	// First, let's try a simple approach - escape all newlines within quoted strings
	
	var result strings.Builder
	inString := false
	escaped := false
	
	for _, char := range jsonStr {
		switch char {
		case '"':
			if !escaped {
				inString = !inString
			}
			result.WriteRune(char)
			escaped = false
		case '\\':
			result.WriteRune(char)
			escaped = !escaped
		case '\n':
			if inString && !escaped {
				// Replace actual newline with escaped newline in JSON strings
				result.WriteString("\\n")
			} else {
				result.WriteRune(char)
			}
			escaped = false
		default:
			result.WriteRune(char)
			escaped = false
		}
	}
	
	return result.String()
} 