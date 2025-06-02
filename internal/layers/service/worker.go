package service

import (
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// EmailJobQueuer defines the interface for enqueuing email jobs.
// This decouples services from the concrete worker implementation.
type EmailJobQueuer interface {
	EnqueueJob(job models.EmailJob)
	// Consider adding Start/Stop methods if the queuer also manages worker lifecycle,
	// though typically Start/Stop would be on the concrete worker type itself.
} 