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
		errorText   string
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
			errorText:   "DATABASE_URL is required",
		},
		{
			name: "missing FIREBASE_SERVICE_ACCOUNT",
			envVars: map[string]string{
				"DATABASE_URL":     "postgres://localhost/test",
				"FIREBASE_API_KEY": "test-api-key",
				"APP_BASE_URL":     "http://localhost:8080",
			},
			expectError: true,
			errorText:   "FIREBASE_SERVICE_ACCOUNT is required",
		},
		{
			name: "missing FIREBASE_API_KEY",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"APP_BASE_URL":             "http://localhost:8080",
			},
			expectError: true,
			errorText:   "FIREBASE_API_KEY is required",
		},
		{
			name: "missing APP_BASE_URL",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"FIREBASE_API_KEY":         "test-api-key",
			},
			expectError: true,
			errorText:   "APP_BASE_URL is required",
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
			errorText:   "invalid PORT",
		},
		{
			name: "default values",
			envVars: map[string]string{
				"DATABASE_URL":             "postgres://localhost/test",
				"FIREBASE_SERVICE_ACCOUNT": `{"type": "service_account"}`,
				"FIREBASE_API_KEY":         "test-api-key",
				"APP_BASE_URL":             "http://localhost:8080",
				// PORT not set should default to 8080
				// SMTP_HOST not set should default to smtp.gmail.com
				// SMTP_PORT not set should default to 587
			},
			expectError: false,
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
				assert.Contains(t, err.Error(), tt.errorText)
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
				
				// Verify default values
				if tt.name == "default values" {
					assert.Equal(t, 8080, config.Port)
					assert.Equal(t, "smtp.gmail.com", config.SMTPHost)
					assert.Equal(t, "587", config.SMTPPort)
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

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_VAR",
			envValue:     "custom_value",
			defaultValue: "default_value",
			expected:     "custom_value",
		},
		{
			name:         "environment variable not set",
			key:          "UNSET_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "environment variable empty string",
			key:          "EMPTY_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.Unsetenv(tt.key)
			
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnvWithDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
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
		"SMTP_HOST",
		"SMTP_PORT",
		"APP_BASE_URL",
		"PORT",
		"RAILWAY_ENVIRONMENT",
	}

	for _, key := range envVars {
		os.Unsetenv(key)
	}
} 