package email

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		expectError       bool
		expectedLogParts  []string
	}{
		// Success cases
		{
			name:             "Valid confirmation email",
			toEmail:          "test@example.com",
			recipientName:    "John Doe",
			confirmationLink: "https://example.com/confirm?token=abc123",
			expectError:      false,
			expectedLogParts: []string{
				"---- SENDING CONFIRMATION EMAIL (CONSOLE MOCK) ----",
				"To: test@example.com",
				"Recipient Name: John Doe",
				"Subject: Confirm Your Subscription",
				"Hi John Doe,",
				"https://example.com/confirm?token=abc123",
				"---- END OF EMAIL (CONSOLE MOCK) ----",
			},
		},
		{
			name:             "Email as recipient name",
			toEmail:          "user@domain.com",
			recipientName:    "user@domain.com",
			confirmationLink: "http://localhost:8080/confirm?token=xyz789",
			expectError:      false,
			expectedLogParts: []string{
				"To: user@domain.com",
				"Recipient Name: user@domain.com",
				"Hi user@domain.com,",
				"http://localhost:8080/confirm?token=xyz789",
			},
		},
		{
			name:             "Unicode recipient name",
			toEmail:          "test@example.com",
			recipientName:    "José María",
			confirmationLink: "https://example.com/confirm?token=unicode123",
			expectError:      false,
			expectedLogParts: []string{
				"To: test@example.com",
				"Recipient Name: José María",
				"Hi José María,",
				"https://example.com/confirm?token=unicode123",
			},
		},
		{
			name:             "Complex confirmation link with query params",
			toEmail:          "test@example.com",
			recipientName:    "Test User",
			confirmationLink: "https://app.example.com/api/subscribers/confirm?token=abc123&newsletter=news-456&utm_source=email",
			expectError:      false,
			expectedLogParts: []string{
				"To: test@example.com",
				"Recipient Name: Test User",
				"https://app.example.com/api/subscribers/confirm?token=abc123&newsletter=news-456&utm_source=email",
			},
		},

		// Edge cases - all should succeed (ConsoleEmailService doesn't validate)
		{
			name:             "Empty email",
			toEmail:          "",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			expectError:      false,
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
			expectError:      false,
			expectedLogParts: []string{
				"To: test@example.com",
				"Recipient Name: ",
				"Hi ,",
			},
		},
		{
			name:             "Empty confirmation link",
			toEmail:          "test@example.com",
			recipientName:    "Test User",
			confirmationLink: "",
			expectError:      false,
			expectedLogParts: []string{
				"To: test@example.com",
				"Recipient Name: Test User",
				"clicking the link below:\n\n",
			},
		},
		{
			name:             "Very long parameters",
			toEmail:          strings.Repeat("a", 100) + "@example.com",
			recipientName:    strings.Repeat("Long Name ", 20),
			confirmationLink: "https://example.com/confirm?token=" + strings.Repeat("x", 200),
			expectError:      false,
			expectedLogParts: []string{
				"To: " + strings.Repeat("a", 100) + "@example.com",
				"Recipient Name: " + strings.Repeat("Long Name ", 20),
				"https://example.com/confirm?token=" + strings.Repeat("x", 200),
			},
		},
		{
			name:             "Special characters in email",
			toEmail:          "test+newsletter@example-domain.co.uk",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			expectError:      false,
			expectedLogParts: []string{
				"To: test+newsletter@example-domain.co.uk",
			},
		},
		{
			name:             "Special characters in name",
			toEmail:          "test@example.com",
			recipientName:    "Test User (Admin) <admin@company.com>",
			confirmationLink: "https://example.com/confirm?token=abc123",
			expectError:      false,
			expectedLogParts: []string{
				"Recipient Name: Test User (Admin) <admin@company.com>",
				"Hi Test User (Admin) <admin@company.com>,",
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

			// Verify error expectation
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

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
		expectError       bool
		expectedLogParts  []string
	}{
		// Success cases
		{
			name:            "Valid newsletter issue",
			toEmail:         "subscriber@example.com",
			recipientName:   "Jane Smith",
			subject:         "Weekly Newsletter #42",
			htmlContent:     "<h1>Welcome to our newsletter!</h1><p>This week's highlights...</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=unsub123",
			expectError:     false,
			expectedLogParts: []string{
				"---- SENDING NEWSLETTER ISSUE (CONSOLE MOCK) ----",
				"To: subscriber@example.com",
				"Recipient Name: Jane Smith",
				"Subject: Weekly Newsletter #42",
				"HTML Content:\n<h1>Welcome to our newsletter!</h1><p>This week's highlights...</p>",
				"Unsubscribe Link: https://example.com/unsubscribe?token=unsub123",
				"---- END OF EMAIL (CONSOLE MOCK) ----",
			},
		},
		{
			name:            "Newsletter with complex HTML content",
			toEmail:         "user@domain.com",
			recipientName:   "User Name",
			subject:         "Monthly Update - December 2024",
			htmlContent:     `<html><body><h1>Monthly Update</h1><div class="content"><p>Hello <strong>subscribers</strong>!</p><ul><li>Item 1</li><li>Item 2</li></ul><a href="https://example.com">Read more</a></div></body></html>`,
			unsubscribeLink: "https://app.example.com/api/subscriptions/unsubscribe?token=monthly123",
			expectError:     false,
			expectedLogParts: []string{
				"To: user@domain.com",
				"Subject: Monthly Update - December 2024",
				"<html><body><h1>Monthly Update</h1>",
				"<strong>subscribers</strong>",
				"https://app.example.com/api/subscriptions/unsubscribe?token=monthly123",
			},
		},
		{
			name:            "Plain text content (no HTML tags)",
			toEmail:         "reader@example.com",
			recipientName:   "Avid Reader",
			subject:         "Simple Newsletter",
			htmlContent:     "This is plain text content without any HTML tags. Just simple newsletter content.",
			unsubscribeLink: "https://example.com/unsubscribe?token=plain123",
			expectError:     false,
			expectedLogParts: []string{
				"To: reader@example.com",
				"Recipient Name: Avid Reader",
				"Subject: Simple Newsletter",
				"HTML Content:\nThis is plain text content without any HTML tags",
				"Unsubscribe Link: https://example.com/unsubscribe?token=plain123",
			},
		},
		{
			name:            "Unicode content and recipient",
			toEmail:         "测试@example.com",
			recipientName:   "张三",
			subject:         "新闻通讯 - Newsletter",
			htmlContent:     "<h1>欢迎 Welcome!</h1><p>This is a multilingual newsletter with 中文 content.</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=unicode123",
			expectError:     false,
			expectedLogParts: []string{
				"To: 测试@example.com",
				"Recipient Name: 张三",
				"Subject: 新闻通讯 - Newsletter",
				"<h1>欢迎 Welcome!</h1>",
				"中文 content",
			},
		},

		// Edge cases - all should succeed (ConsoleEmailService doesn't validate)
		{
			name:            "Empty email",
			toEmail:         "",
			recipientName:   "Test User",
			subject:         "Test Subject",
			htmlContent:     "<p>Test content</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=test123",
			expectError:     false,
			expectedLogParts: []string{
				"To: ",
				"Recipient Name: Test User",
			},
		},
		{
			name:            "Empty recipient name",
			toEmail:         "test@example.com",
			recipientName:   "",
			subject:         "Test Subject",
			htmlContent:     "<p>Test content</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=test123",
			expectError:     false,
			expectedLogParts: []string{
				"To: test@example.com",
				"Recipient Name: ",
			},
		},
		{
			name:            "Empty subject",
			toEmail:         "test@example.com",
			recipientName:   "Test User",
			subject:         "",
			htmlContent:     "<p>Test content</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=test123",
			expectError:     false,
			expectedLogParts: []string{
				"Subject: ",
			},
		},
		{
			name:            "Empty HTML content",
			toEmail:         "test@example.com",
			recipientName:   "Test User",
			subject:         "Test Subject",
			htmlContent:     "",
			unsubscribeLink: "https://example.com/unsubscribe?token=test123",
			expectError:     false,
			expectedLogParts: []string{
				"HTML Content:\n",
				"Unsubscribe Link: https://example.com/unsubscribe?token=test123",
			},
		},
		{
			name:            "Empty unsubscribe link",
			toEmail:         "test@example.com",
			recipientName:   "Test User",
			subject:         "Test Subject",
			htmlContent:     "<p>Test content</p>",
			unsubscribeLink: "",
			expectError:     false,
			expectedLogParts: []string{
				"Unsubscribe Link: ",
			},
		},
		{
			name:            "Very long content",
			toEmail:         "test@example.com",
			recipientName:   "Test User",
			subject:         "Very Long Newsletter Subject " + strings.Repeat("Long ", 50),
			htmlContent:     "<h1>Long Content</h1>" + strings.Repeat("<p>This is a very long paragraph with lots of content. </p>", 100),
			unsubscribeLink: "https://example.com/unsubscribe?token=" + strings.Repeat("x", 500),
			expectError:     false,
			expectedLogParts: []string{
				"Subject: Very Long Newsletter Subject " + strings.Repeat("Long ", 50),
				"<h1>Long Content</h1>",
				"https://example.com/unsubscribe?token=" + strings.Repeat("x", 500),
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

			// Verify error expectation
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify log output contains expected parts
			logOutput := buf.String()
			for _, expectedPart := range tt.expectedLogParts {
				assert.Contains(t, logOutput, expectedPart, "Log output should contain: %s", expectedPart)
			}
		})
	}
}

// TestConsoleEmailService_RealWorldUsagePatterns tests patterns used in actual application
func TestConsoleEmailService_RealWorldUsagePatterns(t *testing.T) {
	service := &ConsoleEmailService{}

	t.Run("Subscriber confirmation flow", func(t *testing.T) {
		// This mimics how SubscriberService uses the email service
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		// Simulate subscriber service calling confirmation email
		email := "newuser@example.com"
		recipientName := email // SubscriberService uses email as recipient name
		confirmationLink := "http://localhost:8080/api/subscribers/confirm?token=conf123"

		err := service.SendConfirmationEmail(email, recipientName, confirmationLink)
		require.NoError(t, err)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "To: newuser@example.com")
		assert.Contains(t, logOutput, "Recipient Name: newuser@example.com")
		assert.Contains(t, logOutput, "http://localhost:8080/api/subscribers/confirm?token=conf123")
	})

	t.Run("Newsletter publishing flow", func(t *testing.T) {
		// This mimics how PublishingService uses the email service
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		// Simulate publishing service sending newsletter
		email := "subscriber@example.com"
		recipientName := email // PublishingService uses email as recipient name
		subject := "Weekly Tech Newsletter #15"
		htmlContent := "<h1>This Week in Tech</h1><p>Latest updates...</p>"
		unsubscribeLink := "http://localhost:8080/api/subscriptions/unsubscribe?token=unsub456"

		err := service.SendNewsletterIssue(email, recipientName, subject, htmlContent, unsubscribeLink)
		require.NoError(t, err)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "To: subscriber@example.com")
		assert.Contains(t, logOutput, "Subject: Weekly Tech Newsletter #15")
		assert.Contains(t, logOutput, "<h1>This Week in Tech</h1>")
		assert.Contains(t, logOutput, "http://localhost:8080/api/subscriptions/unsubscribe?token=unsub456")
	})

	t.Run("Multiple emails in sequence", func(t *testing.T) {
		// Test sending multiple emails (like publishing to multiple subscribers)
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		subscribers := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
		subject := "Newsletter Issue #1"
		content := "<h1>Welcome!</h1><p>First newsletter issue</p>"

		for i, email := range subscribers {
			unsubscribeLink := "http://localhost:8080/api/subscriptions/unsubscribe?token=token" + string(rune('1'+i))
			err := service.SendNewsletterIssue(email, email, subject, content, unsubscribeLink)
			require.NoError(t, err)
		}

		logOutput := buf.String()
		// Verify all emails were logged
		for _, email := range subscribers {
			assert.Contains(t, logOutput, "To: "+email)
		}
		// Should have 3 newsletter sending blocks
		assert.Equal(t, 3, strings.Count(logOutput, "---- SENDING NEWSLETTER ISSUE (CONSOLE MOCK) ----"))
		assert.Equal(t, 3, strings.Count(logOutput, "---- END OF EMAIL (CONSOLE MOCK) ----"))
	})
}

// TestConsoleEmailService_LogOutputFormat tests the exact log format
func TestConsoleEmailService_LogOutputFormat(t *testing.T) {
	service := &ConsoleEmailService{}

	t.Run("Confirmation email log format", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		err := service.SendConfirmationEmail("test@example.com", "Test User", "https://example.com/confirm?token=abc123")
		require.NoError(t, err)

		logOutput := buf.String()
		
		// The actual format uses single log.Printf calls, so we check for content presence
		// rather than exact line-by-line structure
		expectedParts := []string{
			"---- SENDING CONFIRMATION EMAIL (CONSOLE MOCK) ----",
			"To: test@example.com",
			"Recipient Name: Test User (Note: name might not be available for subscriber, using email or generic greeting)",
			"Subject: Confirm Your Subscription",
			"Body (template placeholder):",
			"Hi Test User,",
			"Please confirm your subscription by clicking the link below:",
			"https://example.com/confirm?token=abc123",
			"Thanks!",
			"---- END OF EMAIL (CONSOLE MOCK) ----",
		}

		// Verify all expected parts are present in the log output
		for _, part := range expectedParts {
			assert.Contains(t, logOutput, part, "Log output should contain: %s", part)
		}
		
		// Verify the structure: confirmation email creates 12 lines due to embedded newlines
		// (start, to, recipient, subject, body with embedded newlines, end)
		logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
		assert.Equal(t, 12, len(logLines), "Should have exactly 12 log lines")
		
		// Verify key lines have timestamp prefixes (log.Printf adds timestamps)
		timestampLines := []int{0, 1, 2, 3, 4, 11} // Lines that should have timestamps
		for _, lineIdx := range timestampLines {
			assert.Regexp(t, `^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `, logLines[lineIdx], 
				"Line %d should start with timestamp", lineIdx+1)
		}
	})

	t.Run("Newsletter issue log format", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		err := service.SendNewsletterIssue("subscriber@example.com", "Subscriber Name", "Newsletter Subject", "<h1>Content</h1>", "https://example.com/unsubscribe?token=xyz")
		require.NoError(t, err)

		logOutput := buf.String()
		
		expectedParts := []string{
			"---- SENDING NEWSLETTER ISSUE (CONSOLE MOCK) ----",
			"To: subscriber@example.com",
			"Recipient Name: Subscriber Name",
			"Subject: Newsletter Subject",
			"HTML Content:",
			"<h1>Content</h1>",
			"Unsubscribe Link: https://example.com/unsubscribe?token=xyz",
			"---- END OF EMAIL (CONSOLE MOCK) ----",
		}

		// Verify all expected parts are present in the log output
		for _, part := range expectedParts {
			assert.Contains(t, logOutput, part, "Log output should contain: %s", part)
		}
		
		// Verify the structure: newsletter issue creates 8 lines due to embedded newlines
		// (start, to, recipient, subject, content with embedded newlines, unsubscribe, end)
		logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
		assert.Equal(t, 8, len(logLines), "Should have exactly 8 log lines")
		
		// Verify key lines have timestamp prefixes (log.Printf adds timestamps)
		timestampLines := []int{0, 1, 2, 3, 4, 6, 7} // Lines that should have timestamps
		for _, lineIdx := range timestampLines {
			assert.Regexp(t, `^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `, logLines[lineIdx], 
				"Line %d should start with timestamp", lineIdx+1)
		}
	})
} 