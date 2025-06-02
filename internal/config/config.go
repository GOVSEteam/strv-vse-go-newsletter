package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application
type Config struct {
	// Database configuration
	DatabaseURL       string
	DatabasePublicURL string

	// Firebase configuration
	FirebaseServiceAccount string
	FirebaseAPIKey         string

	// Email configuration
	GoogleAppPassword string
	EmailFrom         string

	// Application configuration
	AppBaseURL string
	Port       int

	// Environment
	RailwayEnvironment string
}

// ConfigError represents a configuration-related error
type ConfigError struct {
	Field   string
	Message string
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("config error: %s - %s", e.Field, e.Message)
}

// Load reads configuration from environment variables and validates required fields
func Load() (*Config, error) {
	// Try to load .env file (optional, might not exist in production)
	_ = godotenv.Load()

	config := &Config{
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		DatabasePublicURL:      os.Getenv("DATABASE_PUBLIC_URL"),
		FirebaseServiceAccount: os.Getenv("FIREBASE_SERVICE_ACCOUNT"),
		FirebaseAPIKey:         os.Getenv("FIREBASE_API_KEY"),
		GoogleAppPassword:      os.Getenv("GOOGLE_APP_PASSWORD"),
		EmailFrom:              os.Getenv("EMAIL_FROM"),
		AppBaseURL:             os.Getenv("APP_BASE_URL"),
		RailwayEnvironment:     os.Getenv("RAILWAY_ENVIRONMENT"),
	}

	// Parse port with default fallback
	portStr := os.Getenv("PORT")
	if portStr == "" {
		config.Port = 8080 // Default port
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, ConfigError{
				Field:   "PORT",
				Message: fmt.Sprintf("invalid port value '%s', must be a valid integer", portStr),
			}
		}
		config.Port = port
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks that all required configuration fields are present
func (c *Config) Validate() error {
	// Check required database configuration
	if c.DatabaseURL == "" {
		return ConfigError{
			Field:   "DATABASE_URL",
			Message: "database URL is required",
		}
	}

	// Check required Firebase configuration
	if c.FirebaseServiceAccount == "" {
		return ConfigError{
			Field:   "FIREBASE_SERVICE_ACCOUNT",
			Message: "Firebase service account JSON is required",
		}
	}

	if c.FirebaseAPIKey == "" {
		return ConfigError{
			Field:   "FIREBASE_API_KEY",
			Message: "Firebase API key is required",
		}
	}

	// Check required email configuration (only if Google App Password is set)
	if c.GoogleAppPassword != "" && c.EmailFrom == "" {
		return ConfigError{
			Field:   "EMAIL_FROM",
			Message: "email from address is required when Google App Password is configured",
		}
	}

	// Check required application configuration
	if c.AppBaseURL == "" {
		return ConfigError{
			Field:   "APP_BASE_URL",
			Message: "application base URL is required",
		}
	}

	return nil
}

// GetDatabaseURL returns the appropriate database URL based on environment
func (c *Config) GetDatabaseURL() string {
	// Use public URL if available and not in Railway environment
	if c.RailwayEnvironment == "" && c.DatabasePublicURL != "" {
		return c.DatabasePublicURL
	}
	return c.DatabaseURL
}

// IsEmailEnabled returns true if email sending is configured
func (c *Config) IsEmailEnabled() bool {
	return c.GoogleAppPassword != "" && c.EmailFrom != ""
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.RailwayEnvironment != ""
} 