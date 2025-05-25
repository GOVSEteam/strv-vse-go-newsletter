package email

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmailServiceSelection tests the logic used in router.go for selecting email service
func TestEmailServiceSelection(t *testing.T) {
	tests := []struct {
		name                string
		resendAPIKey        string
		emailFrom           string
		expectedServiceType string
		expectError         bool
	}{
		{
			name:                "Resend service when API key is set",
			resendAPIKey:        "re_test_api_key_123",
			emailFrom:           "noreply@example.com",
			expectedServiceType: "*email.ResendService",
			expectError:         false,
		},
		{
			name:                "Console service when API key is empty",
			resendAPIKey:        "",
			emailFrom:           "noreply@example.com",
			expectedServiceType: "*email.ConsoleEmailService",
			expectError:         false,
		},
		{
			name:                "Console service when API key is not set",
			resendAPIKey:        "", // Will be unset
			emailFrom:           "",
			expectedServiceType: "*email.ConsoleEmailService",
			expectError:         false,
		},
		{
			name:                "Resend service fails when EMAIL_FROM missing",
			resendAPIKey:        "re_test_api_key_456",
			emailFrom:           "",
			expectedServiceType: "",
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.resendAPIKey != "" {
				t.Setenv("RESEND_API_KEY", tt.resendAPIKey)
			} else {
				os.Unsetenv("RESEND_API_KEY")
			}
			
			if tt.emailFrom != "" {
				t.Setenv("EMAIL_FROM", tt.emailFrom)
			} else {
				os.Unsetenv("EMAIL_FROM")
			}

			// Simulate router.go email service selection logic
			var emailSvc EmailService
			var err error
			
			if os.Getenv("RESEND_API_KEY") != "" {
				emailSvc, err = NewResendService()
				if err != nil && !tt.expectError {
					t.Fatalf("Unexpected error creating Resend service: %v", err)
				}
			} else {
				emailSvc = NewConsoleEmailService()
			}

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, emailSvc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, emailSvc)
				
				// Verify service type
				serviceType := getServiceTypeName(emailSvc)
				assert.Equal(t, tt.expectedServiceType, serviceType)
			}
		})
	}
}

// Helper function to get service type name for testing
func getServiceTypeName(service EmailService) string {
	switch service.(type) {
	case *ConsoleEmailService:
		return "*email.ConsoleEmailService"
	case *ResendService:
		return "*email.ResendService"
	default:
		return "unknown"
	}
}

// TestEmailServiceInterfaceConsistency tests that both services implement the interface consistently
func TestEmailServiceInterfaceConsistency(t *testing.T) {
	// Test ConsoleEmailService
	consoleService := NewConsoleEmailService()
	var _ EmailService = consoleService

	// Test ResendService (with valid config)
	t.Setenv("RESEND_API_KEY", "test_key")
	t.Setenv("EMAIL_FROM", "test@example.com")
	
	resendService, err := NewResendService()
	require.NoError(t, err)
	var _ EmailService = resendService

	// Both services should have the same interface methods
	services := []EmailService{consoleService, resendService}
	
	for i, service := range services {
		serviceName := []string{"ConsoleEmailService", "ResendService"}[i]
		
		t.Run(serviceName+" interface methods", func(t *testing.T) {
			// Test that methods exist and can be called (even if they fail)
			err1 := service.SendConfirmationEmail("test@example.com", "Test User", "https://example.com/confirm")
			err2 := service.SendNewsletterIssue("test@example.com", "Test User", "Subject", "<p>Content</p>", "https://example.com/unsubscribe")
			
			// ConsoleEmailService should never error
			if serviceName == "ConsoleEmailService" {
				assert.NoError(t, err1)
				assert.NoError(t, err2)
			}
			// ResendService will error without proper setup, but methods should exist
			// (We're just testing interface consistency here)
		})
	}
}

// TestEmailServiceEnvironmentHandling tests how services handle different environment configurations
func TestEmailServiceEnvironmentHandling(t *testing.T) {
	t.Run("Environment variable isolation", func(t *testing.T) {
		// Test that environment changes don't affect already created services
		t.Setenv("RESEND_API_KEY", "initial_key")
		t.Setenv("EMAIL_FROM", "initial@example.com")
		
		service1, err := NewResendService()
		require.NoError(t, err)
		
		resendSvc1 := service1.(*ResendService)
		assert.Equal(t, "initial@example.com", resendSvc1.fromEmail)
		
		// Change environment
		t.Setenv("EMAIL_FROM", "changed@example.com")
		
		service2, err := NewResendService()
		require.NoError(t, err)
		
		resendSvc2 := service2.(*ResendService)
		assert.Equal(t, "changed@example.com", resendSvc2.fromEmail)
		
		// Original service should be unchanged
		assert.Equal(t, "initial@example.com", resendSvc1.fromEmail)
	})

	t.Run("Environment variable precedence", func(t *testing.T) {
		// Test that environment variables take precedence
		t.Setenv("RESEND_API_KEY", "env_key")
		t.Setenv("EMAIL_FROM", "env@example.com")
		
		service, err := NewResendService()
		require.NoError(t, err)
		
		resendSvc := service.(*ResendService)
		assert.Equal(t, "env@example.com", resendSvc.fromEmail)
		assert.NotNil(t, resendSvc.client)
	})
}

// TestEmailServiceErrorPropagation tests how errors propagate through the service layer
func TestEmailServiceErrorPropagation(t *testing.T) {
	t.Run("ConsoleEmailService error propagation", func(t *testing.T) {
		service := NewConsoleEmailService()
		
		// ConsoleEmailService should never return errors
		err1 := service.SendConfirmationEmail("", "", "")
		err2 := service.SendNewsletterIssue("", "", "", "", "")
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})

	t.Run("ResendService constructor error propagation", func(t *testing.T) {
		// Test missing API key
		os.Unsetenv("RESEND_API_KEY")
		t.Setenv("EMAIL_FROM", "test@example.com")
		
		service, err := NewResendService()
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "RESEND_API_KEY environment variable not set")
		
		// Test missing EMAIL_FROM
		t.Setenv("RESEND_API_KEY", "test_key")
		os.Unsetenv("EMAIL_FROM")
		
		service, err = NewResendService()
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "EMAIL_FROM environment variable not set")
	})
}

// TestEmailServiceUsagePatterns tests real-world usage patterns from the application
func TestEmailServiceUsagePatterns(t *testing.T) {
	t.Run("SubscriberService pattern with ConsoleEmailService", func(t *testing.T) {
		// This mimics how SubscriberService would use the email service
		emailService := NewConsoleEmailService()
		
		// Simulate subscription confirmation flow
		subscriberEmail := "newuser@example.com"
		recipientName := subscriberEmail // SubscriberService uses email as name
		confirmationToken := "conf_token_123"
		confirmationLink := "http://localhost:8080/api/subscribers/confirm?token=" + confirmationToken
		
		err := emailService.SendConfirmationEmail(subscriberEmail, recipientName, confirmationLink)
		assert.NoError(t, err)
	})

	t.Run("PublishingService pattern with ConsoleEmailService", func(t *testing.T) {
		// This mimics how PublishingService would use the email service
		emailService := NewConsoleEmailService()
		
		// Simulate newsletter publishing flow
		subscribers := []struct {
			email           string
			unsubscribeToken string
		}{
			{"user1@example.com", "unsub_token_1"},
			{"user2@example.com", "unsub_token_2"},
			{"user3@example.com", "unsub_token_3"},
		}
		
		subject := "Weekly Newsletter #15"
		htmlContent := "<h1>This Week's Updates</h1><p>Newsletter content...</p>"
		appBaseURL := "http://localhost:8080"
		
		for _, subscriber := range subscribers {
			unsubscribeLink := appBaseURL + "/api/subscriptions/unsubscribe?token=" + subscriber.unsubscribeToken
			recipientName := subscriber.email // PublishingService uses email as name
			
			err := emailService.SendNewsletterIssue(subscriber.email, recipientName, subject, htmlContent, unsubscribeLink)
			assert.NoError(t, err)
		}
	})

	t.Run("Service layer error handling pattern", func(t *testing.T) {
		// Test how service layer should handle email service errors
		
		// With ConsoleEmailService (never errors)
		consoleService := NewConsoleEmailService()
		err := consoleService.SendConfirmationEmail("test@example.com", "Test User", "https://example.com/confirm")
		assert.NoError(t, err)
		// Service layer can proceed without error handling
		
		// With ResendService (can error)
		t.Setenv("RESEND_API_KEY", "test_key")
		t.Setenv("EMAIL_FROM", "test@example.com")
		
		resendService, err := NewResendService()
		require.NoError(t, err)
		
		// ResendService will error when actually trying to send (no real API)
		// But constructor succeeds, so service layer can handle send errors appropriately
		assert.NotNil(t, resendService)
	})
}

// TestEmailServiceConfiguration tests various configuration scenarios
func TestEmailServiceConfiguration(t *testing.T) {
	t.Run("Development configuration", func(t *testing.T) {
		// Typical development setup - no RESEND_API_KEY
		os.Unsetenv("RESEND_API_KEY")
		
		// Should fall back to ConsoleEmailService
		var emailSvc EmailService
		if os.Getenv("RESEND_API_KEY") != "" {
			emailSvc, _ = NewResendService()
		} else {
			emailSvc = NewConsoleEmailService()
		}
		
		assert.NotNil(t, emailSvc)
		assert.IsType(t, &ConsoleEmailService{}, emailSvc)
	})

	t.Run("Production configuration", func(t *testing.T) {
		// Typical production setup - RESEND_API_KEY set
		t.Setenv("RESEND_API_KEY", "re_prod_key_123")
		t.Setenv("EMAIL_FROM", "noreply@company.com")
		
		var emailSvc EmailService
		var err error
		if os.Getenv("RESEND_API_KEY") != "" {
			emailSvc, err = NewResendService()
			require.NoError(t, err)
		} else {
			emailSvc = NewConsoleEmailService()
		}
		
		assert.NotNil(t, emailSvc)
		assert.IsType(t, &ResendService{}, emailSvc)
	})

	t.Run("Staging configuration", func(t *testing.T) {
		// Staging might use real email service but with different FROM address
		t.Setenv("RESEND_API_KEY", "re_staging_key_456")
		t.Setenv("EMAIL_FROM", "staging-noreply@company.com")
		
		emailSvc, err := NewResendService()
		require.NoError(t, err)
		
		resendSvc := emailSvc.(*ResendService)
		assert.Equal(t, "staging-noreply@company.com", resendSvc.fromEmail)
	})

	t.Run("Invalid configuration handling", func(t *testing.T) {
		// Test graceful handling of invalid configurations
		
		// Invalid: RESEND_API_KEY set but EMAIL_FROM missing
		t.Setenv("RESEND_API_KEY", "re_key_123")
		os.Unsetenv("EMAIL_FROM")
		
		_, err := NewResendService()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EMAIL_FROM environment variable not set")
		
		// Application should fall back to ConsoleEmailService
		fallbackService := NewConsoleEmailService()
		assert.NotNil(t, fallbackService)
	})
}

// TestEmailServiceConcurrency tests that email services can be used concurrently
func TestEmailServiceConcurrency(t *testing.T) {
	t.Run("ConsoleEmailService concurrent usage", func(t *testing.T) {
		service := NewConsoleEmailService()
		
		// Test concurrent access (should be safe since it's stateless)
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func(id int) {
				err1 := service.SendConfirmationEmail("test@example.com", "Test User", "https://example.com/confirm")
				err2 := service.SendNewsletterIssue("test@example.com", "Test User", "Subject", "<p>Content</p>", "https://example.com/unsubscribe")
				
				assert.NoError(t, err1)
				assert.NoError(t, err2)
				done <- true
			}(i)
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("ResendService concurrent creation", func(t *testing.T) {
		// Test that multiple ResendService instances can be created concurrently
		t.Setenv("RESEND_API_KEY", "test_key")
		t.Setenv("EMAIL_FROM", "test@example.com")
		
		done := make(chan EmailService, 5)
		
		for i := 0; i < 5; i++ {
			go func() {
				service, err := NewResendService()
				assert.NoError(t, err)
				done <- service
			}()
		}
		
		// Collect all services
		services := make([]EmailService, 5)
		for i := 0; i < 5; i++ {
			services[i] = <-done
			assert.NotNil(t, services[i])
		}
		
		// All should be valid ResendService instances
		for _, service := range services {
			assert.IsType(t, &ResendService{}, service)
		}
	})
} 