package service

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// EmailService defines an interface for sending emails.
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// --- ConsoleEmailService ---

// ConsoleEmailService is an implementation of EmailService that logs emails to the console.
// Useful for development and testing.
type ConsoleEmailService struct {
	logger *log.Logger
}

// NewConsoleEmailService creates a new ConsoleEmailService.
// If logger is nil, log.Default() (which writes to os.Stderr) will be used.
func NewConsoleEmailService(logger *log.Logger) *ConsoleEmailService {
	if logger == nil {
		logger = log.New(os.Stderr, "[ConsoleEmailService] ", log.LstdFlags)
	}
	return &ConsoleEmailService{logger: logger}
}

// SendEmail logs the email details to the console.
func (s *ConsoleEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	select {
	case <-ctx.Done():
		s.logger.Printf("Context cancelled before sending email to %s: %v", to, ctx.Err())
		return ctx.Err()
	default:
	}

	s.logger.Printf("---- SENDING EMAIL (CONSOLE MOCK) ----")
	s.logger.Printf("To: %s", to)
	s.logger.Printf("Subject: %s", subject)
	s.logger.Printf("Body:\n%s", body)
	s.logger.Printf("---- END OF EMAIL (CONSOLE MOCK) ----")
	return nil
}

// --- GmailEmailService ---

// GmailEmailServiceConfig holds configuration for the GmailEmailService.
// These values would typically be populated from the main application configuration.
type GmailEmailServiceConfig struct {
	From     string // Sender's email address (e.g., from config.EmailFrom)
	Password string // App Password for Gmail (e.g., from config.GoogleAppPassword)
	SMTPHost string // e.g., "smtp.gmail.com" (should be in main config)
	SMTPPort string // e.g., "587" (should be in main config)
}

// GmailEmailService sends emails using an SMTP server (configured for Gmail in mind).
type GmailEmailService struct {
	config GmailEmailServiceConfig
	logger *log.Logger
}

// NewGmailEmailService creates a new GmailEmailService.
// If logger is nil, a default logger writing to os.Stderr will be used.
func NewGmailEmailService(config GmailEmailServiceConfig, logger *log.Logger) (*GmailEmailService, error) {
	if config.From == "" || config.Password == "" || config.SMTPHost == "" || config.SMTPPort == "" {
		return nil, fmt.Errorf("gmail email service: missing required configuration fields (From, Password, SMTPHost, SMTPPort)")
	}
	if logger == nil {
		logger = log.New(os.Stderr, "[GmailEmailService] ", log.LstdFlags)
	}
	return &GmailEmailService{config: config, logger: logger}, nil
}

// SendEmail sends an email using SMTP.
// The body is sent as plain text. For HTML emails, the body string should include
// appropriate MIME headers.
func (s *GmailEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	select {
	case <-ctx.Done():
		s.logger.Printf("Context cancelled before sending email to %s: %v", to, ctx.Err())
		return ctx.Err()
	default:
	}

	auth := smtp.PlainAuth("", s.config.From, s.config.Password, s.config.SMTPHost)
	smtpAddr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)

	// Construct the email message.
	// For HTML, the body should start with "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	// For this generic service, we assume the body is pre-formatted if HTML is intended,
	// or it's plain text.
	msg := []byte(fmt.Sprintf("From: %s\r\n"+ // Added From header
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.config.From, to, subject, body))

	s.logger.Printf("Attempting to send email from %s to %s via %s", s.config.From, to, smtpAddr)

	err := smtp.SendMail(smtpAddr, auth, s.config.From, []string{to}, msg)
	if err != nil {
		s.logger.Printf("Failed to send email to %s: %v", to, err)
		return fmt.Errorf("smtp.SendMail (from: %s, to: %s) failed: %w", s.config.From, to, err)
	}

	s.logger.Printf("Email sent successfully from %s to %s", s.config.From, to)
	return nil
} 