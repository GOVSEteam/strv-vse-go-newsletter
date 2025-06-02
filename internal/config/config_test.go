package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorField  string
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"FIREBASE_API_KEY":         "test-api-key",
				"APP_BASE_URL":             "http://localhost:8080",
				"PORT":                     "3000",
			},
			expectError: false,
		},
		{
			name: "missing DATABASE_URL",
			envVars: map[string]string{
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"FIREBASE_API_KEY":         "test-api-key",
				"APP_BASE_URL":             "http://localhost:8080",
			},
			expectError: true,
			errorField:  "DATABASE_URL",
		},
		{
			name: "missing FIREBASE_SERVICE_ACCOUNT",
			envVars: map[string]string{
				"DATABASE_URL":     "postgres://localhost/test",
				"FIREBASE_API_KEY": "test-api-key",
				"APP_BASE_URL":     "http://localhost:8080",
			},
			expectError: true,
			errorField:  "FIREBASE_SERVICE_ACCOUNT",
		},
		{
			name: "missing FIREBASE_API_KEY",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"APP_BASE_URL":             "http://localhost:8080",
			},
			expectError: true,
			errorField:  "FIREBASE_API_KEY",
		},
		{
			name: "missing APP_BASE_URL",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"FIREBASE_API_KEY":         "test-api-key",
			},
			expectError: true,
			errorField:  "APP_BASE_URL",
		},
		{
			name: "invalid PORT",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"FIREBASE_API_KEY":         "test-api-key",
				"APP_BASE_URL":             "http://localhost:8080",
				"PORT":                     "invalid",
			},
			expectError: true,
			errorField:  "PORT",
		},
		{
			name: "email configuration validation",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"FIREBASE_API_KEY":         "test-api-key",
				"APP_BASE_URL":             "http://localhost:8080",
				"GOOGLE_APP_PASSWORD":      "password",
				// Missing EMAIL_FROM should trigger validation error
			},
			expectError: true,
			errorField:  "EMAIL_FROM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			config, err := Load()

			if tt.expectError {
				require.Error(t, err)
				var configErr ConfigError
				require.ErrorAs(t, err, &configErr)
				assert.Equal(t, tt.errorField, configErr.Field)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, config)
				
				// Verify specific values for valid configuration
				if tt.name == "valid configuration" {
					assert.Equal(t, "postgres://localhost/test", config.DatabaseURL)
					assert.Equal(t, `{"type": "service_account"}`, config.FirebaseServiceAccount)
					assert.Equal(t, "test-api-key", config.FirebaseAPIKey)
					assert.Equal(t, "http://localhost:8080", config.AppBaseURL)
					assert.Equal(t, 3000, config.Port)
				}
			}

			// Clean up
			clearEnv()
		})
	}
}

func TestConfig_GetDatabaseURL(t *testing.T) {
	tests := []struct {
		name               string
		databaseURL        string
		databasePublicURL  string
		railwayEnvironment string
		expected           string
	}{
		{
			name:               "use public URL when not in Railway",
			databaseURL:        "postgres://private/db",
			databasePublicURL:  "postgres://public/db",
			railwayEnvironment: "",
			expected:           "postgres://public/db",
		},
		{
			name:               "use private URL in Railway environment",
			databaseURL:        "postgres://private/db",
			databasePublicURL:  "postgres://public/db",
			railwayEnvironment: "production",
			expected:           "postgres://private/db",
		},
		{
			name:               "use private URL when public not available",
			databaseURL:        "postgres://private/db",
			databasePublicURL:  "",
			railwayEnvironment: "",
			expected:           "postgres://private/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				DatabaseURL:        tt.databaseURL,
				DatabasePublicURL:  tt.databasePublicURL,
				RailwayEnvironment: tt.railwayEnvironment,
			}

			result := config.GetDatabaseURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_IsEmailEnabled(t *testing.T) {
	tests := []struct {
		name              string
		googleAppPassword string
		emailFrom         string
		expected          bool
	}{
		{
			name:              "email enabled",
			googleAppPassword: "password",
			emailFrom:         "test@example.com",
			expected:          true,
		},
		{
			name:              "missing password",
			googleAppPassword: "",
			emailFrom:         "test@example.com",
			expected:          false,
		},
		{
			name:              "missing email from",
			googleAppPassword: "password",
			emailFrom:         "",
			expected:          false,
		},
		{
			name:              "both missing",
			googleAppPassword: "",
			emailFrom:         "",
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				GoogleAppPassword: tt.googleAppPassword,
				EmailFrom:         tt.emailFrom,
			}

			result := config.IsEmailEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name               string
		railwayEnvironment string
		expected           bool
	}{
		{
			name:               "production environment",
			railwayEnvironment: "production",
			expected:           true,
		},
		{
			name:               "development environment",
			railwayEnvironment: "",
			expected:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				RailwayEnvironment: tt.railwayEnvironment,
			}

			result := config.IsProduction()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigError_Error(t *testing.T) {
	err := ConfigError{
		Field:   "TEST_FIELD",
		Message: "test message",
	}

	expected := "config error: TEST_FIELD - test message"
	assert.Equal(t, expected, err.Error())
}

// clearEnv removes all environment variables used by the configuration
func clearEnv() {
	envVars := []string{
		"DATABASE_URL",
		"DATABASE_PUBLIC_URL",
		"FIREBASE_SERVICE_ACCOUNT",
		"FIREBASE_API_KEY",
		"GOOGLE_APP_PASSWORD",
		"EMAIL_FROM",
		"APP_BASE_URL",
		"PORT",
		"RAILWAY_ENVIRONMENT",
	}

	for _, key := range envVars {
		os.Unsetenv(key)
	}
} 