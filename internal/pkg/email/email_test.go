package email

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConsoleEmailService_InterfaceCompliance verifies that ConsoleEmailService implements EmailService interface
func TestConsoleEmailService_InterfaceCompliance(t *testing.T) {
	var _ EmailService = &ConsoleEmailService{}
	var _ EmailService = NewConsoleEmailService()
}

// TestNewConsoleEmailService tests the constructor
func TestNewConsoleEmailService(t *testing.T) {
	service := NewConsoleEmailService()
	
	assert.NotNil(t, service)
	assert.IsType(t, &ConsoleEmailService{}, service)
}

// TestConsoleEmailService_SendConfirmationEmail tests the SendConfirmationEmail method
func TestConsoleEmailService_SendConfirmationEmail(t *testing.T) {
	tests := []struct {
		name              string
		toEmail           string
		recipientName     string
		confirmationLink  string
		expectedLogParts  []string
	}{
		{
			name:             "Valid confirmation email",
			toEmail:          "test@example.com",
			recipientName:    "John Doe",
			confirmationLink: "https://example.com/confirm?token=abc123",
			expectedLogParts: []string{
				"---- SENDING CONFIRMATION EMAIL (CONSOLE MOCK) ----",
				"To: test@example.com",
				"Recipient Name: John Doe",
				"Subject:Your Newsletter Subscription Is Active",
				"Hi John Doe,",
				"https://example.com/confirm?token=abc123",
				"---- END OF EMAIL (CONSOLE MOCK) ----",
			},
		},
		{
			name:             "Empty email",
			toEmail:          "",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			expectedLogParts: []string{
				"To: ",
				"Recipient Name: Test User",
			},
		},
		{
			name:             "Empty recipient name",
			toEmail:          "test@example.com",
			recipientName:    "",
			confirmationLink: "https://example.com/confirm?token=abc123",
			expectedLogParts: []string{
				"To: test@example.com",
				"Recipient Name: ",
				"Hi ,",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr) // Restore default output

			service := &ConsoleEmailService{}
			err := service.SendConfirmationEmail(tt.toEmail, tt.recipientName, tt.confirmationLink)

			assert.NoError(t, err)

			// Verify log output contains expected parts
			logOutput := buf.String()
			for _, expectedPart := range tt.expectedLogParts {
				assert.Contains(t, logOutput, expectedPart, "Log output should contain: %s", expectedPart)
			}
		})
	}
}

// TestConsoleEmailService_SendNewsletterIssue tests the SendNewsletterIssue method
func TestConsoleEmailService_SendNewsletterIssue(t *testing.T) {
	tests := []struct {
		name              string
		toEmail           string
		recipientName     string
		subject           string
		htmlContent       string
		unsubscribeLink   string
		expectedLogParts  []string
	}{
		{
			name:            "Valid newsletter issue",
			toEmail:         "subscriber@example.com",
			recipientName:   "Jane Smith",
			subject:         "Weekly Newsletter #1",
			htmlContent:     "<h1>Welcome</h1><p>This is our first newsletter issue.</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=xyz789",
			expectedLogParts: []string{
				"---- SENDING NEWSLETTER ISSUE (CONSOLE MOCK) ----",
				"To: subscriber@example.com",
				"Recipient Name: Jane Smith",
				"Subject: Weekly Newsletter #1",
				"HTML Content:",
				"<h1>Welcome</h1><p>This is our first newsletter issue.</p>",
				"Unsubscribe Link: https://example.com/unsubscribe?token=xyz789",
				"---- END OF EMAIL (CONSOLE MOCK) ----",
			},
		},
		{
			name:            "Empty parameters",
			toEmail:         "",
			recipientName:   "",
			subject:         "",
			htmlContent:     "",
			unsubscribeLink: "",
			expectedLogParts: []string{
				"To: ",
				"Recipient Name: ",
				"Subject: ",
				"HTML Content:",
				"Unsubscribe Link: ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr) // Restore default output

			service := &ConsoleEmailService{}
			err := service.SendNewsletterIssue(tt.toEmail, tt.recipientName, tt.subject, tt.htmlContent, tt.unsubscribeLink)

			assert.NoError(t, err)

			// Verify log output contains expected parts
			logOutput := buf.String()
			for _, expectedPart := range tt.expectedLogParts {
				assert.Contains(t, logOutput, expectedPart, "Log output should contain: %s", expectedPart)
			}
		})
	}
}

// TestConsoleEmailService_BasicUsage tests basic usage patterns
func TestConsoleEmailService_BasicUsage(t *testing.T) {
	service := NewConsoleEmailService()
	
	// Test confirmation email
	err := service.SendConfirmationEmail("user@example.com", "Test User", "https://example.com/confirm")
	assert.NoError(t, err)
	
	// Test newsletter issue
	err = service.SendNewsletterIssue("user@example.com", "Test User", "Subject", "<p>Content</p>", "https://example.com/unsubscribe")
	assert.NoError(t, err)
} 