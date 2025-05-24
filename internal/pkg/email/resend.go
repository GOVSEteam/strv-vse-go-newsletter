package email

import (
	"errors"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

// ResendService implements EmailService using the Resend API.
type ResendService struct {
	client   *resend.Client
	fromEmail string
}

// NewResendService creates a new ResendService.
// It expects RESEND_API_KEY and EMAIL_FROM to be set in environment variables.
func NewResendService() (EmailService, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, errors.New("RESEND_API_KEY environment variable not set")
	}
	fromEmail := os.Getenv("EMAIL_FROM")
	if fromEmail == "" {
		return nil, errors.New("EMAIL_FROM environment variable not set")
	}

	client := resend.NewClient(apiKey)
	return &ResendService{
		client:   client,
		fromEmail: fromEmail,
	}, nil
}

// SendConfirmationEmail sends a subscription confirmation email using Resend.
func (s *ResendService) SendConfirmationEmail(toEmail, recipientName, confirmationLink string) error {
	subject := "Confirm Your Subscription"
	// In a real application, you'd use HTML templates
	htmlBody := fmt.Sprintf(`
		<h1>Hi %s,</h1>
		<p>Please confirm your subscription to our newsletter by clicking the link below:</p>
		<p><a href="%s">Confirm Subscription</a></p>
		<p>If you did not request this, please ignore this email.</p>
		<p>Thanks!</p>
	`, recipientName, confirmationLink)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email via Resend: %w", err)
	}
	if sent.Id == "" {
		return errors.New("failed to send confirmation email via Resend, no ID returned")
	}
	// log.Printf("Confirmation email sent to %s, ID: %s", toEmail, sent.Id)
	return nil
}

// SendNewsletterIssue sends a newsletter issue email using Resend.
func (s *ResendService) SendNewsletterIssue(toEmail, recipientName, subject, htmlContent string) error {
	// Basic HTML structure, ideally use templates
	fullHtmlContent := fmt.Sprintf(`
		<html>
		<head>
			<title>%s</title>
		</head>
		<body>
			<p>Hi %s,</p>
			%s
			<hr>
			<p><small>You are receiving this email because you subscribed to our newsletter.</small></p>
			<p><small><a href="[UNSUBSCRIBE_LINK_PLACEHOLDER]">Unsubscribe</a></small></p> 
		</body>
		</html>
	`, subject, recipientName, htmlContent) // Note: UNSUBSCRIBE_LINK_PLACEHOLDER needs to be replaced

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    fullHtmlContent,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send newsletter issue via Resend: %w", err)
	}
	if sent.Id == "" {
		return errors.New("failed to send newsletter issue via Resend, no ID returned")
	}
	// log.Printf("Newsletter issue '%s' sent to %s, ID: %s", subject, toEmail, sent.Id)
	return nil
}
