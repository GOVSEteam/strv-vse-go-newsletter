package setup

import (
	"fmt"
	"log"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

// NewGmailEmailService creates and configures a Gmail SMTP email service.
// It validates all required configuration and returns a ready-to-use service.
// This follows the same pattern as other setup functions for consistency.
func NewGmailEmailService(from, password, smtpHost, smtpPort string, logger *log.Logger) (service.EmailService, error) {
	if from == "" {
		return nil, fmt.Errorf("email from address is required")
	}
	if password == "" {
		return nil, fmt.Errorf("email password is required")
	}
	if smtpHost == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	if smtpPort == "" {
		return nil, fmt.Errorf("SMTP port is required")
	}

	config := service.GmailEmailServiceConfig{
		From:     from,
		Password: password,
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
	}

	emailService, err := service.NewGmailEmailService(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gmail email service: %w", err)
	}

	return emailService, nil
} 