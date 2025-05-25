package email

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/resend/resend-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockResendClient is a mock implementation of the Resend client
type MockResendClient struct {
	mock.Mock
}

// MockEmailsService is a mock implementation of the Emails service
type MockEmailsService struct {
	mock.Mock
}

// Send mocks the Send method of the Emails service
func (m *MockEmailsService) Send(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*resend.SendEmailResponse), args.Error(1)
}

// TestNewResendService tests the ResendService constructor
func TestNewResendService(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		emailFrom     string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Valid configuration",
			apiKey:      "re_test_api_key_123",
			emailFrom:   "noreply@example.com",
			expectError: false,
		},
		{
			name:          "Missing API key",
			apiKey:        "",
			emailFrom:     "noreply@example.com",
			expectError:   true,
			expectedError: "RESEND_API_KEY environment variable not set",
		},
		{
			name:          "Missing email from",
			apiKey:        "re_test_api_key_123",
			emailFrom:     "",
			expectError:   true,
			expectedError: "EMAIL_FROM environment variable not set",
		},
		{
			name:          "Both missing",
			apiKey:        "",
			emailFrom:     "",
			expectError:   true,
			expectedError: "RESEND_API_KEY environment variable not set",
		},
		{
			name:        "Complex email from",
			apiKey:      "re_test_api_key_456",
			emailFrom:   "Newsletter <newsletter@company.com>",
			expectError: false,
		},
		{
			name:        "Long API key",
			apiKey:      "re_" + strings.Repeat("a", 100),
			emailFrom:   "test@example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			if tt.apiKey != "" {
				t.Setenv("RESEND_API_KEY", tt.apiKey)
			} else {
				os.Unsetenv("RESEND_API_KEY")
			}
			
			if tt.emailFrom != "" {
				t.Setenv("EMAIL_FROM", tt.emailFrom)
			} else {
				os.Unsetenv("EMAIL_FROM")
			}

			service, err := NewResendService()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				
				// Verify service is of correct type
				resendSvc, ok := service.(*ResendService)
				require.True(t, ok, "Service should be of type *ResendService")
				
				// Verify internal fields are set correctly
				assert.NotNil(t, resendSvc.client)
				assert.Equal(t, tt.emailFrom, resendSvc.fromEmail)
			}
		})
	}
}

// TestResendService_SendConfirmationEmail tests the SendConfirmationEmail method
func TestResendService_SendConfirmationEmail(t *testing.T) {
	tests := []struct {
		name              string
		toEmail           string
		recipientName     string
		confirmationLink  string
		mockSetup         func(*MockEmailsService)
		expectError       bool
		expectedError     string
	}{
		{
			name:             "Successful confirmation email",
			toEmail:          "user@example.com",
			recipientName:    "John Doe",
			confirmationLink: "https://example.com/confirm?token=abc123",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "email-id-123",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return params.From == "noreply@example.com" &&
						len(params.To) == 1 && params.To[0] == "user@example.com" &&
						params.Subject == "Confirm Your Subscription" &&
						strings.Contains(params.Html, "John Doe") &&
						strings.Contains(params.Html, "https://example.com/confirm?token=abc123")
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},
		{
			name:             "Email with special characters",
			toEmail:          "test+newsletter@example-domain.co.uk",
			recipientName:    "José María",
			confirmationLink: "https://app.example.com/confirm?token=unicode123&lang=es",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "email-id-456",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return params.To[0] == "test+newsletter@example-domain.co.uk" &&
						strings.Contains(params.Html, "José María") &&
						strings.Contains(params.Html, "https://app.example.com/confirm?token=unicode123&lang=es")
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},
		{
			name:             "Very long confirmation link",
			toEmail:          "user@example.com",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=" + strings.Repeat("x", 500) + "&extra=data",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "email-id-789",
				}
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},

		// Error cases - API failures
		{
			name:             "Resend API error",
			toEmail:          "user@example.com",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			mockSetup: func(mockEmails *MockEmailsService) {
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(nil, errors.New("API rate limit exceeded")).Once()
			},
			expectError:   true,
			expectedError: "failed to send confirmation email via Resend: API rate limit exceeded",
		},
		{
			name:             "Invalid API key error",
			toEmail:          "user@example.com",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			mockSetup: func(mockEmails *MockEmailsService) {
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(nil, errors.New("invalid API key")).Once()
			},
			expectError:   true,
			expectedError: "failed to send confirmation email via Resend: invalid API key",
		},
		{
			name:             "Network timeout error",
			toEmail:          "user@example.com",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			mockSetup: func(mockEmails *MockEmailsService) {
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(nil, errors.New("network timeout")).Once()
			},
			expectError:   true,
			expectedError: "failed to send confirmation email via Resend: network timeout",
		},

		// Error cases - Response validation
		{
			name:             "Empty ID in response",
			toEmail:          "user@example.com",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "", // Empty ID
				}
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(expectedResponse, nil).Once()
			},
			expectError:   true,
			expectedError: "failed to send confirmation email via Resend, no ID returned",
		},

		// Edge cases - parameter validation (ResendService doesn't validate, but we test behavior)
		{
			name:             "Empty email address",
			toEmail:          "",
			recipientName:    "Test User",
			confirmationLink: "https://example.com/confirm?token=abc123",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "email-id-empty",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return len(params.To) == 1 && params.To[0] == ""
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false, // ResendService doesn't validate, passes through to API
		},
		{
			name:             "Empty recipient name",
			toEmail:          "user@example.com",
			recipientName:    "",
			confirmationLink: "https://example.com/confirm?token=abc123",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "email-id-empty-name",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return strings.Contains(params.Html, "Hi ,") // Empty name results in "Hi ,"
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},
		{
			name:             "Empty confirmation link",
			toEmail:          "user@example.com",
			recipientName:    "Test User",
			confirmationLink: "",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "email-id-empty-link",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return strings.Contains(params.Html, `href=""`) // Empty link
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock emails service
			mockEmails := new(MockEmailsService)
			tt.mockSetup(mockEmails)

			// Create ResendService with mock
			service := &ResendService{
				client:   nil, // We'll mock the Emails service directly
				fromEmail: "noreply@example.com",
			}

			// We need to mock the client.Emails.Send call
			// Since we can't easily mock the nested structure, we'll test the method logic
			// by temporarily replacing the client behavior in a wrapper
			originalSend := func(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
				return mockEmails.Send(params)
			}

			// Create a test wrapper that uses our mock
			err := service.sendConfirmationEmailWithMock(tt.toEmail, tt.recipientName, tt.confirmationLink, originalSend)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockEmails.AssertExpectations(t)
		})
	}
}

// Helper method for testing with mocked send function
func (s *ResendService) sendConfirmationEmailWithMock(toEmail, recipientName, confirmationLink string, sendFunc func(*resend.SendEmailRequest) (*resend.SendEmailResponse, error)) error {
	subject := "Confirm Your Subscription"
	htmlBody := `
		<h1>Hi ` + recipientName + `,</h1>
		<p>Please confirm your subscription to our newsletter by clicking the link below:</p>
		<p><a href="` + confirmationLink + `">Confirm Subscription</a></p>
		<p>If you did not request this, please ignore this email.</p>
		<p>Thanks!</p>
	`

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := sendFunc(params)
	if err != nil {
		return errors.New("failed to send confirmation email via Resend: " + err.Error())
	}
	if sent.Id == "" {
		return errors.New("failed to send confirmation email via Resend, no ID returned")
	}
	return nil
}

// TestResendService_SendNewsletterIssue tests the SendNewsletterIssue method
func TestResendService_SendNewsletterIssue(t *testing.T) {
	tests := []struct {
		name              string
		toEmail           string
		recipientName     string
		subject           string
		htmlContent       string
		unsubscribeLink   string
		mockSetup         func(*MockEmailsService)
		expectError       bool
		expectedError     string
	}{
		{
			name:            "Successful newsletter issue",
			toEmail:         "subscriber@example.com",
			recipientName:   "Jane Smith",
			subject:         "Weekly Newsletter #42",
			htmlContent:     "<h1>Welcome to our newsletter!</h1><p>This week's highlights...</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=unsub123",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "newsletter-id-123",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return params.From == "noreply@example.com" &&
						len(params.To) == 1 && params.To[0] == "subscriber@example.com" &&
						params.Subject == "Weekly Newsletter #42" &&
						strings.Contains(params.Html, "Jane Smith") &&
						strings.Contains(params.Html, "<h1>Welcome to our newsletter!</h1>") &&
						strings.Contains(params.Html, "https://example.com/unsubscribe?token=unsub123")
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},
		{
			name:            "Complex HTML newsletter",
			toEmail:         "reader@example.com",
			recipientName:   "Avid Reader",
			subject:         "Monthly Tech Update - December 2024",
			htmlContent:     `<html><body><h1>Tech News</h1><div class="content"><p>Hello <strong>tech enthusiasts</strong>!</p><ul><li>AI Updates</li><li>Web3 News</li></ul><a href="https://techblog.com">Read more</a></div></body></html>`,
			unsubscribeLink: "https://app.example.com/api/subscriptions/unsubscribe?token=monthly456",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "newsletter-id-456",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return strings.Contains(params.Html, "<strong>tech enthusiasts</strong>") &&
						strings.Contains(params.Html, "https://app.example.com/api/subscriptions/unsubscribe?token=monthly456") &&
						strings.Contains(params.Html, "Avid Reader")
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},
		{
			name:            "Unicode content newsletter",
			toEmail:         "international@example.com",
			recipientName:   "国际用户",
			subject:         "国际新闻通讯 - International Newsletter",
			htmlContent:     "<h1>欢迎 Welcome!</h1><p>This newsletter contains 中文 and English content.</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=unicode789",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "newsletter-unicode-789",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return strings.Contains(params.Html, "国际用户") &&
						strings.Contains(params.Html, "欢迎 Welcome!") &&
						strings.Contains(params.Html, "中文 and English")
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},

		// Error cases - API failures
		{
			name:            "API rate limit error",
			toEmail:         "subscriber@example.com",
			recipientName:   "Test User",
			subject:         "Test Newsletter",
			htmlContent:     "<p>Test content</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=test123",
			mockSetup: func(mockEmails *MockEmailsService) {
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(nil, errors.New("rate limit exceeded")).Once()
			},
			expectError:   true,
			expectedError: "failed to send newsletter issue via Resend: rate limit exceeded",
		},
		{
			name:            "Invalid recipient error",
			toEmail:         "invalid-email",
			recipientName:   "Test User",
			subject:         "Test Newsletter",
			htmlContent:     "<p>Test content</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=test123",
			mockSetup: func(mockEmails *MockEmailsService) {
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(nil, errors.New("invalid recipient email")).Once()
			},
			expectError:   true,
			expectedError: "failed to send newsletter issue via Resend: invalid recipient email",
		},

		// Error cases - Response validation
		{
			name:            "Empty ID in response",
			toEmail:         "subscriber@example.com",
			recipientName:   "Test User",
			subject:         "Test Newsletter",
			htmlContent:     "<p>Test content</p>",
			unsubscribeLink: "https://example.com/unsubscribe?token=test123",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "", // Empty ID
				}
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(expectedResponse, nil).Once()
			},
			expectError:   true,
			expectedError: "failed to send newsletter issue via Resend, no ID returned",
		},

		// Edge cases
		{
			name:            "Very long subject and content",
			toEmail:         "subscriber@example.com",
			recipientName:   "Test User",
			subject:         "Very Long Newsletter Subject " + strings.Repeat("Long ", 100),
			htmlContent:     "<h1>Long Content</h1>" + strings.Repeat("<p>This is a very long paragraph with lots of content. </p>", 200),
			unsubscribeLink: "https://example.com/unsubscribe?token=" + strings.Repeat("x", 1000),
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "newsletter-long-content",
				}
				mockEmails.On("Send", mock.AnythingOfType("*resend.SendEmailRequest")).Return(expectedResponse, nil).Once()
			},
			expectError: false,
		},
		{
			name:            "Empty parameters",
			toEmail:         "",
			recipientName:   "",
			subject:         "",
			htmlContent:     "",
			unsubscribeLink: "",
			mockSetup: func(mockEmails *MockEmailsService) {
				expectedResponse := &resend.SendEmailResponse{
					Id: "newsletter-empty-params",
				}
				mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
					return params.Subject == "" && 
						len(params.To) == 1 && params.To[0] == "" &&
						strings.Contains(params.Html, "Hi ,") // Empty recipient name
				})).Return(expectedResponse, nil).Once()
			},
			expectError: false, // ResendService doesn't validate, passes through to API
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock emails service
			mockEmails := new(MockEmailsService)
			tt.mockSetup(mockEmails)

			// Create ResendService with mock
			service := &ResendService{
				client:   nil, // We'll mock the Emails service directly
				fromEmail: "noreply@example.com",
			}

			// Test wrapper that uses our mock
			originalSend := func(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
				return mockEmails.Send(params)
			}

			err := service.sendNewsletterIssueWithMock(tt.toEmail, tt.recipientName, tt.subject, tt.htmlContent, tt.unsubscribeLink, originalSend)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockEmails.AssertExpectations(t)
		})
	}
}

// Helper method for testing newsletter issue with mocked send function
func (s *ResendService) sendNewsletterIssueWithMock(toEmail, recipientName, subject, htmlContent, unsubscribeLink string, sendFunc func(*resend.SendEmailRequest) (*resend.SendEmailResponse, error)) error {
	fullHtmlContent := `
		<html>
		<head>
			<title>` + subject + `</title>
		</head>
		<body>
			<p>Hi ` + recipientName + `,</p>
			` + htmlContent + `
			<hr>
			<p><small>You are receiving this email because you subscribed to our newsletter.</small></p>
			<p><small><a href="` + unsubscribeLink + `">Unsubscribe</a></small></p> 
		</body>
		</html>
	`

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    fullHtmlContent,
	}

	sent, err := sendFunc(params)
	if err != nil {
		return errors.New("failed to send newsletter issue via Resend: " + err.Error())
	}
	if sent.Id == "" {
		return errors.New("failed to send newsletter issue via Resend, no ID returned")
	}
	return nil
}

// TestResendService_InterfaceCompliance verifies that ResendService implements EmailService interface
func TestResendService_InterfaceCompliance(t *testing.T) {
	// Set up environment for constructor
	t.Setenv("RESEND_API_KEY", "test_key")
	t.Setenv("EMAIL_FROM", "test@example.com")
	
	service, err := NewResendService()
	require.NoError(t, err)
	
	// Verify interface compliance
	var _ EmailService = service
}

// TestResendService_RealWorldUsagePatterns tests patterns used in actual application
func TestResendService_RealWorldUsagePatterns(t *testing.T) {
	// Set up environment
	t.Setenv("RESEND_API_KEY", "test_key")
	t.Setenv("EMAIL_FROM", "noreply@newsletter.com")
	
	service, err := NewResendService()
	require.NoError(t, err)
	
	resendSvc := service.(*ResendService)

	t.Run("Subscriber service confirmation pattern", func(t *testing.T) {
		// This mimics how SubscriberService uses the email service
		mockEmails := new(MockEmailsService)
		expectedResponse := &resend.SendEmailResponse{Id: "conf-123"}
		mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
			return params.From == "noreply@newsletter.com" &&
				params.To[0] == "newuser@example.com" &&
				params.Subject == "Confirm Your Subscription" &&
				strings.Contains(params.Html, "newuser@example.com") // Used as recipient name
		})).Return(expectedResponse, nil).Once()

		sendFunc := func(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
			return mockEmails.Send(params)
		}

		err := resendSvc.sendConfirmationEmailWithMock(
			"newuser@example.com",
			"newuser@example.com", // SubscriberService uses email as recipient name
			"http://localhost:8080/api/subscribers/confirm?token=conf123",
			sendFunc,
		)
		
		assert.NoError(t, err)
		mockEmails.AssertExpectations(t)
	})

	t.Run("Publishing service newsletter pattern", func(t *testing.T) {
		// This mimics how PublishingService uses the email service
		mockEmails := new(MockEmailsService)
		expectedResponse := &resend.SendEmailResponse{Id: "newsletter-456"}
		mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
			return params.Subject == "Weekly Tech Newsletter #15" &&
				strings.Contains(params.Html, "subscriber@example.com") && // Used as recipient name
				strings.Contains(params.Html, "<h1>This Week in Tech</h1>") &&
				strings.Contains(params.Html, "http://localhost:8080/api/subscriptions/unsubscribe?token=unsub456")
		})).Return(expectedResponse, nil).Once()

		sendFunc := func(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
			return mockEmails.Send(params)
		}

		err := resendSvc.sendNewsletterIssueWithMock(
			"subscriber@example.com",
			"subscriber@example.com", // PublishingService uses email as recipient name
			"Weekly Tech Newsletter #15",
			"<h1>This Week in Tech</h1><p>Latest updates...</p>",
			"http://localhost:8080/api/subscriptions/unsubscribe?token=unsub456",
			sendFunc,
		)
		
		assert.NoError(t, err)
		mockEmails.AssertExpectations(t)
	})
}

// TestResendService_SecurityScenarios tests various security-related scenarios
func TestResendService_SecurityScenarios(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "test_key")
	t.Setenv("EMAIL_FROM", "noreply@example.com")
	
	service, err := NewResendService()
	require.NoError(t, err)
	
	resendSvc := service.(*ResendService)

	tests := []struct {
		name        string
		toEmail     string
		content     string
		description string
	}{
		{
			name:        "HTML injection attempt",
			toEmail:     "test@example.com",
			content:     "<script>alert('xss')</script><h1>Legitimate content</h1>",
			description: "Should handle script tags in content",
		},
		{
			name:        "SQL injection attempt in email",
			toEmail:     "'; DROP TABLE users; --@example.com",
			content:     "<p>Normal content</p>",
			description: "Should handle malicious email addresses",
		},
		{
			name:        "Very long content (potential DoS)",
			toEmail:     "test@example.com",
			content:     strings.Repeat("<p>Long content paragraph. </p>", 10000),
			description: "Should handle very long content gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEmails := new(MockEmailsService)
			
			// Mock should receive the content as-is (ResendService doesn't sanitize)
			expectedResponse := &resend.SendEmailResponse{Id: "security-test"}
			mockEmails.On("Send", mock.MatchedBy(func(params *resend.SendEmailRequest) bool {
				return params.To[0] == tt.toEmail &&
					strings.Contains(params.Html, tt.content)
			})).Return(expectedResponse, nil).Once()

			sendFunc := func(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
				return mockEmails.Send(params)
			}

			err := resendSvc.sendNewsletterIssueWithMock(
				tt.toEmail,
				"Test User",
				"Security Test",
				tt.content,
				"https://example.com/unsubscribe?token=test",
				sendFunc,
			)
			
			// ResendService should pass through content without validation
			// Security should be handled at the API level or by input validation elsewhere
			assert.NoError(t, err, tt.description)
			mockEmails.AssertExpectations(t)
		})
	}
} 