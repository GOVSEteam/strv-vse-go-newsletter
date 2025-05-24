package email

import "log"

// EmailService defines an interface for sending emails.
type EmailService interface {
	SendConfirmationEmail(toEmail, recipientName, confirmationLink string) error
	SendNewsletterIssue(toEmail, recipientName, subject, htmlContent, unsubscribeLink string) error
	// Potentially other email types later, e.g.:
	// SendNewsletter(toEmail string, subject string, htmlBody string, textBody string) error
}

// ConsoleEmailService is an implementation of EmailService that logs emails to the console.
// This is useful for development and testing without sending real emails.
type ConsoleEmailService struct{}

// NewConsoleEmailService creates a new ConsoleEmailService.
func NewConsoleEmailService() EmailService {
	return &ConsoleEmailService{}
}

// SendConfirmationEmail logs the confirmation email details to the console.
func (s *ConsoleEmailService) SendConfirmationEmail(toEmail, recipientName, confirmationLink string) error {
	log.Printf("---- SENDING CONFIRMATION EMAIL (CONSOLE MOCK) ----")
	log.Printf("To: %s", toEmail)
	log.Printf("Recipient Name: %s (Note: name might not be available for subscriber, using email or generic greeting)", recipientName)
	log.Printf("Subject: Confirm Your Subscription")
	log.Printf("Body (template placeholder):\nHi %s,\n\nPlease confirm your subscription by clicking the link below:\n%s\n\nThanks!", recipientName, confirmationLink)
	log.Printf("---- END OF EMAIL (CONSOLE MOCK) ----")
	return nil // Simulate successful send
}

// SendNewsletterIssue logs the newsletter issue email details to the console.
func (s *ConsoleEmailService) SendNewsletterIssue(toEmail, recipientName, subject, htmlContent, unsubscribeLink string) error {
	log.Printf("---- SENDING NEWSLETTER ISSUE (CONSOLE MOCK) ----")
	log.Printf("To: %s", toEmail)
	log.Printf("Recipient Name: %s", recipientName)
	log.Printf("Subject: %s", subject)
	log.Printf("HTML Content:\n%s", htmlContent)
	log.Printf("Unsubscribe Link: %s", unsubscribeLink)
	log.Printf("---- END OF EMAIL (CONSOLE MOCK) ----")
	return nil // Simulate successful send
}
