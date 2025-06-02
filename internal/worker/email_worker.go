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
// It implicitly implements the service.EmailJobQueuer interface.
type EmailWorker struct {
	emailService service.EmailService
	jobChannel   chan models.EmailJob
	wg           sync.WaitGroup
	// quitSignal chan struct{} // For more immediate shutdown signal to workers if needed
	// deadLetterQueue chan models.EmailJob // For jobs that fail repeatedly
}

// NewEmailWorker creates a new EmailWorker.
// bufferSize determines the capacity of the job channel.
func NewEmailWorker(emailService service.EmailService, bufferSize int) *EmailWorker {
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}
	return &EmailWorker{
		emailService: emailService,
		jobChannel:   make(chan models.EmailJob, bufferSize),
		// quitSignal: make(chan struct{}),
	}
}

// EnqueueJob adds an email job to the worker's queue.
// This method makes EmailWorker implement service.EmailJobQueuer.
func (ew *EmailWorker) EnqueueJob(job models.EmailJob) {
	// In a production system, you might want to handle the case where the channel is full.
	// For example, log a warning, drop the job, or return an error.
	// For simplicity, this will block if the channel is full.
	// Also, consider checking for a shutdown signal before enqueuing if ew.quitSignal is used.
	select {
	case ew.jobChannel <- job:
		log.Printf("Enqueued email job for: %s, Subject: %s", job.To, job.Subject)
	// case <-ew.quitSignal: // If using a quitSignal for immediate shutdown
	// 	log.Printf("Worker is shutting down, not enqueuing job for: %s", job.To)
	default:
		// This case can be reached if the jobChannel is full and non-blocking sends are preferred.
		// To make it blocking as intended by the original code, the default case can be removed
		// or this select can be removed and ew.jobChannel <- job directly used.
		// For this refactor, sticking to the original blocking behavior if channel is full:
	ew.jobChannel <- job
		log.Printf("Enqueued email job (potentially blocked until space): %s, Subject: %s", job.To, job.Subject)
	}
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
		// Pass a unique ID to each worker for better logging.
		go ew.worker(ctx, i+1)
	}

	// Goroutine to listen for context cancellation from the main app context
	// This is one way to signal shutdown. Another is a dedicated Stop method that closes quitSignal.
	go func() {
		<-ctx.Done() // Wait for the context to be cancelled
		log.Printf("Main context cancelled. Initiating email worker shutdown...")
		ew.Stop() // Call Stop to close jobChannel and wait for workers
	}()
}

// worker is the actual goroutine that processes jobs from the jobChannel.
func (ew *EmailWorker) worker(mainCtx context.Context, id int) {
	defer ew.wg.Done()
	log.Printf("Email worker goroutine %d started", id)

	for {
		select {
		// case <-ew.quitSignal: // If using a quitSignal for immediate shutdown
		// 	log.Printf("Email worker goroutine %d received quit signal. Exiting.", id)
		// 	return
		case <-mainCtx.Done(): // Listen to the main context passed to Start
			log.Printf("Email worker goroutine %d: main context cancelled. Draining remaining jobs and exiting.", id)
			// Drain any remaining jobs. This loop will exit when jobChannel is closed and empty.
			for job := range ew.jobChannel {
				ew.processJob(id, job)
			}
			log.Printf("Email worker goroutine %d finished draining jobs. Exiting.", id)
			return
		case job, ok := <-ew.jobChannel:
			if !ok {
				// jobChannel has been closed, worker should terminate.
				log.Printf("Email worker goroutine %d: job channel closed. Exiting.", id)
				return
			}
			ew.processJob(id, job)
		}
	}
}

func (ew *EmailWorker) processJob(workerID int, job models.EmailJob) {
	log.Printf("Worker %d: Processing email job for NewsletterID: %s, To: %s, Subject: %s", workerID, job.NewsletterID, job.To, job.Subject)
	
	// Create a new context for each email sending attempt, possibly with a timeout.
	// Using context.Background() as a base for the send operation context.
	sendCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 30-second timeout for sending an email
	defer cancel()                                                              // Ensure cancel is called to free resources

	err := ew.emailService.SendEmail(sendCtx, job.To, job.Subject, job.Body)
	if err != nil {
		log.Printf("Worker %d: Failed to send email to %s for NewsletterID %s (Subject: %s): %v", workerID, job.To, job.NewsletterID, job.Subject, err)
		// TODO: Implement dead letter queue or retry logic here.
		// For now, just logging the error.
	} else {
		log.Printf("Worker %d: Successfully sent email to %s for NewsletterID %s (Subject: %s)", workerID, job.To, job.NewsletterID, job.Subject)
	}
}

// Stop initiates a graceful shutdown of the email worker.
// It closes the job channel and waits for all worker goroutines to finish processing.
func (ew *EmailWorker) Stop() {
	log.Println("Email worker Stop() called. Closing job channel...")
	// TODO: Ensure Stop() is idempotent and thread-safe if called multiple times.
	// Consider using sync.Once for closing channels if Stop can be called concurrently.
	
	// close(ew.quitSignal) // If using quitSignal, close it here to signal workers immediately.
	
	close(ew.jobChannel) // Close the channel to signal workers no more new jobs will come.
	                     // Workers will process remaining jobs and then exit when channel is empty and closed.
	
	ew.wg.Wait() // Wait for all worker goroutines to finish their current work and exit.
	log.Println("All email worker goroutines have finished. Shutdown complete.")
}

// TODO: Implement retry logic and dead letter queue (DLQ)
// A simple DLQ could be another channel where jobs that fail repeatedly are sent.
// Another process could then monitor the DLQ for manual inspection or alerts.

// TODO: Consider adding more sophisticated logging and metrics
// (e.g., number of jobs processed, failed, average processing time).

// TODO: Ensure thread safety if multiple goroutines can call EnqueueJob or Stop concurrently.
// For EnqueueJob, channel sends are generally thread-safe.
// For Stop, closing channels and wg.Wait() needs care if Stop itself can be called multiple times. 