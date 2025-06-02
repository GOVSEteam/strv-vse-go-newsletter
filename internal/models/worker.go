package models

// EmailJob defines the data for an email to be sent by a worker.
type EmailJob struct {
	To           string
	Subject      string
	Body         string
	NewsletterID string // For logging/tracking, and potentially for specific email templates or unsubscribe links
	// Consider adding fields like:
	// Attempts int // Number of times this job has been attempted
	// MaxAttempts int // Maximum number of retries for this job
	// Priority int // If a priority queue is used
	// Type string // e.g., "newsletter_issue", "confirmation_email" if the worker handles multiple types
} 