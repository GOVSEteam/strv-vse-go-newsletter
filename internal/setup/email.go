package setup

import (
	"fmt"
	"log"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

// NewGmailEmailService creates and configures a Gmail SMTP email service.
// It validates all required configuration and returns a ready-to-use service.
// This follows the same pattern as other setup functions for consistency.
func NewGmailEmailService(from, password, smtpHost, smtpPort string, logger *log.Logger) (service.EmailService, error) {
	if from == "" {
		return nil, fmt.Errorf("email from address is required: %w", apperrors.ErrValidation)
	}
	if password == "" {
		return nil, fmt.Errorf("email password is required: %w", apperrors.ErrValidation)
	}
	if smtpHost == "" {
		return nil, fmt.Errorf("SMTP host is required: %w", apperrors.ErrValidation)
	}
	if smtpPort == "" {
		return nil, fmt.Errorf("SMTP port is required: %w", apperrors.ErrValidation)
	}

	config := service.GmailEmailServiceConfig{
		From:     from,
		Password: password,
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
	}

	emailService, err := service.NewGmailEmailService(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gmail email service: %w", apperrors.ErrInternal)
	}

	return emailService, nil
} 