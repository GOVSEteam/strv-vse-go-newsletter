package email

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
)

// GmailService implements EmailService using Gmail's SMTP.
type GmailService struct {
	fromEmail string
	password  string
	smtpHost  string
	smtpPort  string
}

// NewGmailService creates a new GmailService.
// It expects EMAIL_FROM and GMAIL_APP_PASSWORD to be set in environment variables.
func NewGmailService() (EmailService, error) {
	fromEmail := os.Getenv("EMAIL_FROM")
	if fromEmail == "" {
		return nil, errors.New("EMAIL_FROM environment variable not set")
	}
	password := os.Getenv("GOOGLE_APP_PASSWORD")
	if password == "" {
		return nil, errors.New("GOOGLE_APP_PASSWORD environment variable not set")
	}

	return &GmailService{
		fromEmail: fromEmail,
		password:  password,
		smtpHost:  "smtp.gmail.com",
		smtpPort:  "587",
	}, nil
}

// SendConfirmationEmail sends a subscription confirmation email using Gmail's SMTP.
func (s *GmailService) SendConfirmationEmail(toEmail, recipientName, unsubscribeLink string) error {
	subject := "Your Newsletter Subscription Is Active"
	htmlBody := fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Your subscription to our newsletter is now active!</p>
		<p>Thanks!</p>
		<hr>
		<p><small>If you didn't sign up for this or have changed your mind, you can unsubscribe at any time by clicking the link below:</small></p>
		<p><small><a href="%s">Unsubscribe from Newsletter</a></small></p>

	`, recipientName, unsubscribeLink)

	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlBody)

	auth := smtp.PlainAuth("", s.fromEmail, s.password, s.smtpHost)
	err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.fromEmail, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email via Gmail SMTP: %w", err)
	}
	return nil
}

// SendNewsletterIssue sends a newsletter issue email using Gmail's SMTP.
func (s *GmailService) SendNewsletterIssue(toEmail, recipientName, subject, htmlContent, unsubscribeLink string) error {
	fullHtmlContent := fmt.Sprintf(`
		<p>Hi %s,</p>
		%s
		<hr>
		<p><small>You are receiving this email because you subscribed to our newsletter.</small></p>
		<p><small><a href="%s">Unsubscribe</a></small></p> 
	`, recipientName, htmlContent, unsubscribeLink)

	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		fullHtmlContent)

	auth := smtp.PlainAuth("", s.fromEmail, s.password, s.smtpHost)
	err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.fromEmail, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send newsletter issue via Gmail SMTP: %w", err)
	}
	return nil
}
