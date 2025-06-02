package repository

import (
	"context"
	_ "embed" // Required for //go:embed
	"errors"  // For errors.Is
	"fmt"
	"time"

	"github.com/jackc/pgerrcode" // For pg error codes
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

type Newsletter struct {
	ID          string    `json:"id"`
	EditorID    string    `json:"editor_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toModel converts a dbNewsletter to a models.Newsletter domain object.
func (dbNl *dbNewsletter) toModel() models.Newsletter {
	return models.Newsletter{
		ID:          dbNl.ID,
		EditorID:    dbNl.EditorID,
		Name:        dbNl.Name,
		Description: dbNl.Description,
		CreatedAt:   dbNl.CreatedAt,
		UpdatedAt:   dbNl.UpdatedAt,
	}
}

// toModelPtr converts a dbNewsletter to a *models.Newsletter domain object.
func (dbNl *dbNewsletter) toModelPtr() *models.Newsletter {
	m := dbNl.toModel()
	return &m
}

// NewsletterRepository defines the interface for newsletter data access.
type NewsletterRepository interface {
	ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error)
	CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error)
	GetNewsletterByIDAndEditorID(ctx context.Context, newsletterID string, editorID string) (*models.Newsletter, error)
	UpdateNewsletter(ctx context.Context, newsletterID string, editorID string, name *string, description *string) (*models.Newsletter, error)
	DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error
	GetNewsletterByNameAndEditorID(ctx context.Context, name string, editorID string) (*models.Newsletter, error)
	GetNewsletterByID(ctx context.Context, newsletterID string) (*models.Newsletter, error)
}

// PostgresNewsletterRepo is the PostgreSQL implementation of NewsletterRepository.
type PostgresNewsletterRepo struct {
	db *pgxpool.Pool
}

// NewsletterRepo creates a new PostgresNewsletterRepo.
// Note: This constructor name is kept from the original version for simplicity.
// Consider renaming to NewPostgresNewsletterRepo for consistency if preferred.
func NewsletterRepo(db *pgxpool.Pool) NewsletterRepository {
	return &PostgresNewsletterRepo{db: db}
}

// ListNewslettersByEditorID fetches a paginated list of newsletters for a specific editor.
func (r *PostgresNewsletterRepo) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error) {
	rows, err := r.db.Query(ctx, listNewslettersByEditorIDQuery, editorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: query: %w", err)
	}
	defer rows.Close()

	var dbNewsletters []dbNewsletter
	for rows.Next() {
		var nl dbNewsletter
		if errScan := rows.Scan(&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt); errScan != nil {
			return nil, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: scan: %w", errScan)
		}
		dbNewsletters = append(dbNewsletters, nl)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: rows error: %w", err)
	}

	newsletters := make([]models.Newsletter, 0, len(dbNewsletters))
	for i := range dbNewsletters {
		newsletters = append(newsletters, dbNewsletters[i].toModel())
	}

	var totalCount int
	err = r.db.QueryRow(ctx, countNewslettersByEditorIDQuery, editorID).Scan(&totalCount)
	if err != nil {
		// Return already fetched newsletters if count fails, but log or wrap error appropriately.
		return newsletters, 0, fmt.Errorf("newsletter repo: ListNewslettersByEditorID: count query: %w", err)
	}

	return newsletters, totalCount, nil
}

// CreateNewsletter creates a new newsletter.
func (r *PostgresNewsletterRepo) CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error) {
	var nl dbNewsletter
	err := r.db.QueryRow(ctx, createNewsletterQuery, editorID, name, description).Scan(
		&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("newsletter repo: CreateNewsletter: %w", apperrors.ErrConflict)
		}
		return nil, fmt.Errorf("newsletter repo: CreateNewsletter: scan: %w", err)
	}
	return nl.toModelPtr(), nil
}

// GetNewsletterByIDAndEditorID fetches a newsletter by its ID and verifies editor ownership.
func (r *PostgresNewsletterRepo) GetNewsletterByIDAndEditorID(newsletterID string, editorID string) (*Newsletter, error) {
	query := `
		SELECT id, editor_id, name, description, created_at, updated_at 
		FROM newsletters 
		WHERE id = $1 AND editor_id = $2`
	row := r.db.QueryRow(query, newsletterID, editorID)

	var n Newsletter
	err := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("newsletter repo: GetNewsletterByIDAndEditorID: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("newsletter repo: GetNewsletterByIDAndEditorID: scan: %w", err)
	}
	return nl.toModelPtr(), nil
}

// UpdateNewsletter updates a newsletter's name and/or description and its updated_at timestamp.
// It ensures that the update is performed by the owner.
func (r *PostgresNewsletterRepo) UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*Newsletter, error) {
	currentNewsletter, err := r.GetNewsletterByIDAndEditorID(newsletterID, editorID)
	if err != nil {
		// GetNewsletterByIDAndEditorID already wraps ErrNoRows to ErrNewsletterNotFound and adds context
		return nil, fmt.Errorf("newsletter repo: UpdateNewsletter: fetching current: %w", err)
	}

	updatedName := currentNewsletter.Name
	if name != nil {
		updatedName = *name
	}

	updatedDescription := currentNewsletter.Description
	if description != nil {
		updatedDescription = *description
	}

	// Ensure name is not empty if provided for update
	if name != nil && *name == "" {
		// This validation should ideally be in the service or handler,
		// but as a safeguard here if name is being explicitly set to empty.
		// For this example, we'll allow it, assuming higher layers validate.
		// If name is mandatory for a newsletter, the table schema should reflect that.
	}

	query := `
		UPDATE newsletters
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3 AND editor_id = $4
		RETURNING id, editor_id, name, description, created_at, updated_at`

	row := r.db.QueryRow(query, updatedName, updatedDescription, newsletterID, editorID)

	var n Newsletter
	errScan := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt)
	if errScan != nil {
		// If the RETURNING clause didn't find a row (e.g., id or editor_id didn't match,
		// which shouldn't happen if GetNewsletterByIDAndEditorID passed),
		// sql.ErrNoRows would be returned here.
		return nil, errScan
	}
	return nl.toModelPtr(), nil
}

// DeleteNewsletter removes a newsletter by its ID, ensuring it belongs to the editor.
func (r *PostgresNewsletterRepo) DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error {
	cmdTag, err := r.db.Exec(ctx, deleteNewsletterQuery, newsletterID, editorID)
	if err != nil {
		return fmt.Errorf("newsletter repo: DeleteNewsletter: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("newsletter repo: DeleteNewsletter: %w", apperrors.ErrNewsletterNotFound)
	}
	return nil
}

// GetNewsletterByNameAndEditorID fetches a newsletter by its name and editor ID.
func (r *PostgresNewsletterRepo) GetNewsletterByNameAndEditorID(ctx context.Context, name string, editorID string) (*models.Newsletter, error) {
	var nl dbNewsletter
	err := r.db.QueryRow(ctx, getNewsletterByNameAndEditorIDQuery, name, editorID).Scan(
		&nl.ID, &nl.EditorID, &nl.Name, &nl.Description, &nl.CreatedAt, &nl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("newsletter repo: GetNewsletterByNameAndEditorID: %w", apperrors.ErrNewsletterNotFound)
		}
		// Check for unique constraint violation if applicable, though less likely for a GetByName.
		// For now, just a general error.
		return nil, fmt.Errorf("newsletter repo: GetNewsletterByNameAndEditorID: scan: %w", err)
	}
	return nl.toModelPtr(), nil
}

// GetNewsletterByID fetches a newsletter by its ID, without checking editor ownership.
func (r *PostgresNewsletterRepo) GetNewsletterByID(newsletterID string) (*Newsletter, error) {
	query := `
		SELECT id, editor_id, name, description, created_at, updated_at 
		FROM newsletters 
		WHERE id = $1`
	row := r.db.QueryRow(query, newsletterID)

	var n Newsletter
	err := row.Scan(&n.ID, &n.EditorID, &n.Name, &n.Description, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("newsletter repo: GetNewsletterByID: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("newsletter repo: GetNewsletterByID: scan: %w", err)
	}
	return nl.toModelPtr(), nil
}
