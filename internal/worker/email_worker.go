package worker

import (
	// Or a structured logger
	"context"
	"log"
	"sync"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // Import for EmailJob
)

// EmailJob defines the data for an email to be sent.
type EmailJob struct {
	To           string
	Subject      string
	Body         string
	NewsletterID string // For logging/tracking purposes
	// Potentially add fields like attempts, lastAttemptAt for retry logic
}

// EmailWorker manages a pool of goroutines that process email jobs from a channel.
type EmailWorker struct {
	emailService service.EmailService
	jobChannel   chan models.EmailJob // Changed to models.EmailJob
	wg           sync.WaitGroup     // To wait for workers to finish during shutdown
	// Potentially add a quit channel for signaling workers to stop
	// Potentially add a deadLetterQueue chan models.EmailJob
}

// NewEmailWorker creates a new EmailWorker.
// bufferSize determines the capacity of the job channel.
func NewEmailWorker(emailService service.EmailService, bufferSize int) *EmailWorker {
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}
	return &EmailWorker{
		emailService: emailService,
		jobChannel:   make(chan models.EmailJob, bufferSize), // Changed to models.EmailJob
		// quitSignal: make(chan struct{}), // Added for graceful shutdown
	}
}

// EnqueueJob adds an email job to the worker's queue.
// It returns an error if the job channel is full (e.g., if not using a buffered channel or it's at capacity and blocking).
// For simplicity with a buffered channel, we'll assume it can block or use a select with a default to drop/log if full.
// Here, we'll let it block if the channel is full.
func (ew *EmailWorker) EnqueueJob(job models.EmailJob) { // Changed to models.EmailJob
	// TODO: Consider adding a select with a timeout or a check for a shutdown signal
	// if ew.isShutdown() { log.Printf("Worker is shutting down, not enqueuing job for: %s", job.To); return }
	ew.jobChannel <- job
	log.Printf("Enqueued email job for: %s, Subject: %s", job.To, job.Subject)
}

// Start launches the email worker goroutines.
// It listens for context cancellation to initiate a graceful shutdown.
func (ew *EmailWorker) Start(ctx context.Context, workerCount int) {
	if workerCount <= 0 {
		workerCount = 5 // Default worker count
	}

	log.Printf("Starting email worker with %d goroutines...", workerCount)

	for i := 0; i < workerCount; i++ {
		ew.wg.Add(1)
		go ew.worker(i + 1) // Pass worker ID for logging
	}

	// Goroutine to listen for context cancellation
	go func() {
		<-ctx.Done()
		log.Printf("Context cancelled. Shutting down email worker...")
		ew.Stop()
	}()
}

// worker is the actual goroutine that processes jobs from the jobChannel.
func (ew *EmailWorker) worker(id int) {
	defer ew.wg.Done()
	log.Printf("Email worker goroutine %d started", id)

	for job := range ew.jobChannel { // job is now models.EmailJob
		log.Printf("Worker %d: Processing email job for NewsletterID: %s, To: %s, Subject: %s", id, job.NewsletterID, job.To, job.Subject)
		// Create a new context for each email sending attempt, possibly with a timeout
		// For now, using the worker's main context (which might be too long-lived or broad)
		// It's better to use context.Background() or a short-lived context for the SendEmail call itself.
		// However, the service.EmailService.SendEmail already takes a context.
		// We need to ensure the context passed to SendEmail is appropriate.
		// Let's assume the service handles its own timeout or uses the passed context.
		// The overall worker context (ctx in Start) is for the worker's lifecycle.

		// TODO: The context passed to SendEmail should ideally be derived from the request context
		// that triggered the email, or a new background context with a specific timeout.
		// For now, we'll pass a background context.
		sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30-second timeout for sending an email

		err := ew.emailService.SendEmail(sendCtx, job.To, job.Subject, job.Body)
		if err != nil {
			log.Printf("Worker %d: Failed to send email to %s for NewsletterID %s (Subject: %s): %v", id, job.To, job.NewsletterID, job.Subject, err)
			// TODO: Implement dead letter queue or retry logic here
			// For now, just logging the error.
		} else {
			log.Printf("Worker %d: Successfully sent email to %s for NewsletterID %s (Subject: %s)", id, job.To, job.NewsletterID, job.Subject)
		}
		cancel() // Release resources associated with sendCtx
	}
	log.Printf("Email worker goroutine %d stopped", id)
}

// Stop initiates a graceful shutdown of the email worker.
// It closes the job channel and waits for all worker goroutines to finish.
func (ew *EmailWorker) Stop() {
	log.Println("Email worker Stop() called. Closing job channel...")
	// TODO: Ensure Stop() is idempotent and thread-safe if called multiple times
	close(ew.jobChannel) // Close the channel to signal workers to stop processing new jobs
	ew.wg.Wait()         // Wait for all worker goroutines to finish
	log.Println("All email worker goroutines have finished. Shutdown complete.")
}

// TODO: Implement retry logic and dead letter queue (DLQ)
// A simple DLQ could be another channel where jobs that fail repeatedly are sent.
// Another process could then monitor the DLQ for manual inspection or alerts.

// TODO: Consider adding more sophisticated logging and metrics
// (e.g., number of jobs processed, failed, average processing time). 