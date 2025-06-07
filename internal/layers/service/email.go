package service

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
	SendConfirmationEmailHTML(ctx context.Context, to, recipientName, unsubscribeLink string) error
	SendNewsletterIssueHTML(ctx context.Context, to, recipientName, subject, body, unsubscribeLink string) error
}

// GmailEmailServiceConfig holds configuration for Gmail SMTP service
type GmailEmailServiceConfig struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort string
}

// GmailEmailService implements EmailService using Gmail SMTP
type GmailEmailService struct {
	config GmailEmailServiceConfig
	logger *log.Logger
}

// NewGmailEmailService creates a new Gmail email service
func NewGmailEmailService(config GmailEmailServiceConfig, logger *log.Logger) (*GmailEmailService, error) {
	if config.From == "" {
		return nil, fmt.Errorf("email from address is required")
	}
	if config.Password == "" {
		return nil, fmt.Errorf("email password is required")
	}
	if config.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	if config.SMTPPort == "" {
		return nil, fmt.Errorf("SMTP port is required")
	}

	return &GmailEmailService{
		config: config,
		logger: logger,
	}, nil
}

// SendEmail sends an email using Gmail SMTP
func (s *GmailEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	auth := smtp.PlainAuth("", s.config.From, s.config.Password, s.config.SMTPHost)
	
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
	
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	
	if err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message)); err != nil {
		s.logger.Printf("Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	s.logger.Printf("Email sent successfully to %s", to)
	return nil
}

// SendConfirmationEmailHTML sends a subscription confirmation email with HTML content
func (s *GmailEmailService) SendConfirmationEmailHTML(ctx context.Context, to, recipientName, unsubscribeLink string) error {
	subject := "Subscription Confirmation"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body>
	<h1>Welcome, %s!</h1>
	<p>Thank you for subscribing to our newsletter.</p>
	<p>You can unsubscribe at any time by clicking <a href="%s">here</a>.</p>
</body>
</html>`, recipientName, unsubscribeLink)
	
	return s.sendHTMLEmail(ctx, to, subject, htmlBody)
}

// SendNewsletterIssueHTML sends a newsletter issue with HTML content
func (s *GmailEmailService) SendNewsletterIssueHTML(ctx context.Context, to, recipientName, subject, body, unsubscribeLink string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body>
	<h1>%s</h1>
	<p>Dear %s,</p>
	<div>%s</div>
	<hr>
	<p><small><a href="%s">Unsubscribe</a></small></p>
</body>
</html>`, subject, recipientName, body, unsubscribeLink)
	
	return s.sendHTMLEmail(ctx, to, subject, htmlBody)
}

// sendHTMLEmail sends an HTML email using Gmail SMTP
func (s *GmailEmailService) sendHTMLEmail(ctx context.Context, to, subject, htmlBody string) error {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	auth := smtp.PlainAuth("", s.config.From, s.config.Password, s.config.SMTPHost)
	
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, htmlBody)
	
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	
	if err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(message)); err != nil {
		s.logger.Printf("Failed to send HTML email to %s: %v", to, err)
		return fmt.Errorf("failed to send HTML email: %w", err)
	}
	
	s.logger.Printf("HTML email sent successfully to %s", to)
	return nil
} 