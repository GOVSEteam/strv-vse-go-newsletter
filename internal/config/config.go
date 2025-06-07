package config

import (
	"encoding/base64"
	"encoding/json"
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
		// First, try to parse the JSON as-is to see if it's already valid
		var testParse map[string]interface{}
		if err := json.Unmarshal([]byte(plainJSON), &testParse); err == nil {
			// JSON is already valid, return as-is
			return plainJSON
		}
		
		// JSON is invalid, check if it contains actual newlines that need escaping
		if strings.Contains(plainJSON, "\n") {
			// Try to fix unescaped newlines in JSON strings
			fixed := fixFirebaseJSONNewlines(plainJSON)
			// Verify the fix worked
			if err := json.Unmarshal([]byte(fixed), &testParse); err == nil {
				return fixed
			}
		}
		
		// If it contains literal \n sequences, try converting them to actual newlines
		if strings.Contains(plainJSON, "\\n") {
			converted := strings.ReplaceAll(plainJSON, "\\n", "\n")
			// Check if this creates valid JSON
			if err := json.Unmarshal([]byte(converted), &testParse); err == nil {
				return converted
			}
		}
		
		// Return the original if nothing worked
		return plainJSON
	}
	
	return ""
}

// fixFirebaseJSONNewlines fixes JSON that contains unescaped newlines, particularly
// in the private_key field of Firebase service account JSON
func fixFirebaseJSONNewlines(jsonStr string) string {
	// More robust approach: properly escape newlines only within JSON string values
	var result strings.Builder
	inString := false
	escaped := false
	
	for _, char := range jsonStr {
		switch char {
		case '\\':
			result.WriteRune(char)
			// Toggle escaped state only if we're in a string
			if inString {
				escaped = !escaped
			}
		case '"':
			if !escaped {
				inString = !inString
			}
			result.WriteRune(char)
			escaped = false
		case '\n':
			if inString && !escaped {
				// We're inside a JSON string and this is an unescaped newline
				// Replace it with the escaped version
				result.WriteString("\\n")
			} else {
				// We're outside a string or it's already escaped, keep as is
				result.WriteRune(char)
			}
			escaped = false
		case '\r':
			if inString && !escaped {
				// Also handle carriage returns
				result.WriteString("\\r")
			} else {
				result.WriteRune(char)
			}
			escaped = false
		case '\t':
			if inString && !escaped {
				// Also handle tabs
				result.WriteString("\\t")
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